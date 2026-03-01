package king

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type controlMessage struct {
	ServiceID string `json:"service_id"`
	Token     string `json:"token"`
}

type controlResponse struct {
	OK bool `json:"ok"`
}

type serviceAuth struct {
	token string
}

type tunnelServer struct {
	bindPort        int
	tlsConfig       *tls.Config
	mu              sync.RWMutex
	services        map[string]serviceAuth // service_id -> auth
	tcpListeners    map[int]net.Listener   // remote_port -> listener
	quicConns       map[string]*quic.Conn  // service_id -> QUIC connection
	onLingConnected func()
}

func newTunnelServer(bindPort int, tlsConfig *tls.Config) *tunnelServer {
	return &tunnelServer{
		bindPort:     bindPort,
		tlsConfig:    tlsConfig,
		services:     make(map[string]serviceAuth),
		tcpListeners: make(map[int]net.Listener),
		quicConns:    make(map[string]*quic.Conn),
	}
}

func (ts *tunnelServer) updateServices(services map[string]serviceAuth) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.services = services
}

func (ts *tunnelServer) run(ctx context.Context) error {
	listener, err := quic.ListenAddr(
		net.JoinHostPort("0.0.0.0", strconv.Itoa(ts.bindPort)),
		ts.tlsConfig,
		&quic.Config{
			KeepAlivePeriod:    5 * time.Second,
			MaxIncomingStreams: 1024,
		},
	)
	if err != nil {
		return err
	}

	slog.Info("QUIC listener started", "bind_port", ts.bindPort)

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for ctx.Err() == nil {
		conn, err := listener.Accept(ctx)
		if err != nil {
			slog.Error("QUIC accept failed", "error", err)
			continue
		}

		go ts.handleConnection(conn)
	}
	return nil
}

func (ts *tunnelServer) handleConnection(conn *quic.Conn) {
	stream, err := conn.AcceptStream(conn.Context())
	if err != nil {
		slog.Error("Failed to accept control stream", "error", err)
		return
	}

	decoder := json.NewDecoder(stream)
	var messages []controlMessage
	if err := decoder.Decode(&messages); err != nil {
		slog.Error("Failed to decode control messages", "error", err)
		_ = stream.Close()
		_ = conn.CloseWithError(1, "invalid control message")
		return
	}

	ts.mu.RLock()
	validServiceIDs := make([]string, 0, len(messages))
	for _, msg := range messages {
		auth, exists := ts.services[msg.ServiceID]
		if !exists || auth.token != msg.Token {
			ts.mu.RUnlock()
			slog.Error("Invalid token for service", "service_id", msg.ServiceID)
			_ = stream.Close()
			_ = conn.CloseWithError(2, "authentication failed")
			return
		}
		validServiceIDs = append(validServiceIDs, msg.ServiceID)
	}
	ts.mu.RUnlock()

	response := controlResponse{OK: true}
	if err := json.NewEncoder(stream).Encode(response); err != nil {
		slog.Error("Failed to send control response", "error", err)
		return
	}

	slog.Info("Ling authenticated", "service_ids", validServiceIDs)

	ts.mu.Lock()
	for _, serviceID := range validServiceIDs {
		ts.quicConns[serviceID] = conn
	}
	ts.mu.Unlock()

	if ts.onLingConnected != nil {
		ts.onLingConnected()
	}

	<-conn.Context().Done()

	ts.mu.Lock()
	for _, serviceID := range validServiceIDs {
		delete(ts.quicConns, serviceID)
	}
	ts.mu.Unlock()

	slog.Info("Ling disconnected", "service_ids", validServiceIDs)
}

func (ts *tunnelServer) ensureTCPListener(remotePort int, serviceID string) {
	ts.mu.Lock()
	if _, exists := ts.tcpListeners[remotePort]; exists {
		ts.mu.Unlock()
		return
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(remotePort)))
	if err != nil {
		ts.mu.Unlock()
		slog.Error("Failed to listen on remote port", "remote_port", remotePort, "error", err)
		return
	}
	ts.tcpListeners[remotePort] = listener
	ts.mu.Unlock()

	slog.Info("TCP listener started", "remote_port", remotePort, "service_id", serviceID)

	go func() {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				return
			}

			go ts.handleTCPConnection(tcpConn, serviceID)
		}
	}()
}

func (ts *tunnelServer) removeTCPListener(remotePort int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if listener, exists := ts.tcpListeners[remotePort]; exists {
		_ = listener.Close()
		delete(ts.tcpListeners, remotePort)
		slog.Info("TCP listener removed", "remote_port", remotePort)
	}
}

func (ts *tunnelServer) handleTCPConnection(tcpConn net.Conn, serviceID string) {
	defer func() { _ = tcpConn.Close() }()

	ts.mu.RLock()
	quicConn, exists := ts.quicConns[serviceID]
	ts.mu.RUnlock()

	if !exists {
		return
	}

	// Use the QUIC connection's context so we detect dead connections
	streamCtx, cancel := context.WithTimeout(quicConn.Context(), 5*time.Second)
	defer cancel()

	stream, err := quicConn.OpenStreamSync(streamCtx)
	if err != nil {
		return
	}

	serviceIDBytes := []byte(serviceID)
	serviceIDLen := len(serviceIDBytes)
	if serviceIDLen > math.MaxUint16 {
		_ = stream.Close()
		return
	}
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(serviceIDLen))
	if _, err := stream.Write(header); err != nil {
		_ = stream.Close()
		return
	}
	if _, err := stream.Write(serviceIDBytes); err != nil {
		_ = stream.Close()
		return
	}

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(stream, tcpConn)
		_ = stream.Close()
		close(done)
	}()
	_, _ = io.Copy(tcpConn, stream)
	stream.CancelRead(0)
	<-done
}

func (ts *tunnelServer) close() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for port, listener := range ts.tcpListeners {
		_ = listener.Close()
		delete(ts.tcpListeners, port)
	}
	for serviceID, conn := range ts.quicConns {
		_ = conn.CloseWithError(0, "server shutting down")
		delete(ts.quicConns, serviceID)
	}
}
