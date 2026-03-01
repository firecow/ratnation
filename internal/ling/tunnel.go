package ling

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type tunnelControlMessage struct {
	ServiceID string `json:"service_id"`
	Token     string `json:"token"`
}

type tunnelControlResponse struct {
	OK bool `json:"ok"`
}

type tunnelService struct {
	serviceID string
	token     string
	localAddr string
}

type tunnelClient struct {
	mu          sync.Mutex
	connections map[string]quic.Connection // king_bind_addr -> connection
	onConnected func()
}

func newTunnelClient() *tunnelClient {
	return &tunnelClient{
		connections: make(map[string]quic.Connection),
	}
}

func (tc *tunnelClient) ensureConnection(ctx context.Context, group *kingGroup, localAddrs map[string]string) {
	kingAddr := net.JoinHostPort(group.host, strconv.Itoa(group.bindPort))

	tc.mu.Lock()
	if conn, exists := tc.connections[kingAddr]; exists {
		select {
		case <-conn.Context().Done():
			delete(tc.connections, kingAddr)
		default:
			tc.mu.Unlock()
			return
		}
	}
	tc.mu.Unlock()

	if group.certPEM == "" {
		slog.Warn("King has no certificate, skipping connection", "king_addr", kingAddr)
		return
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(group.certPEM))

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		ServerName: "burrow",
		NextProtos: []string{"burrow"},
		MinVersion: tls.VersionTLS13,
	}

	conn, err := quic.DialAddr(ctx, kingAddr, tlsConfig, &quic.Config{
		KeepAlivePeriod:    5 * time.Second,
		MaxIncomingStreams: 1024,
	})
	if err != nil {
		slog.Error("Failed to dial king", "king_addr", kingAddr, "error", err)
		return
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		slog.Error("Failed to open control stream", "error", err)
		_ = conn.CloseWithError(1, "failed to open control stream")
		return
	}

	messages := make([]tunnelControlMessage, 0, len(group.services))
	for _, svc := range group.services {
		messages = append(messages, tunnelControlMessage{
			ServiceID: svc.serviceID,
			Token:     svc.token,
		})
	}

	if err := json.NewEncoder(stream).Encode(messages); err != nil {
		slog.Error("Failed to send control messages", "error", err)
		_ = conn.CloseWithError(1, "failed to send control messages")
		return
	}

	var response tunnelControlResponse
	if err := json.NewDecoder(stream).Decode(&response); err != nil {
		slog.Error("Failed to read control response", "error", err)
		_ = conn.CloseWithError(1, "failed to read control response")
		return
	}

	if !response.OK {
		slog.Error("King rejected registration", "king_addr", kingAddr)
		_ = conn.CloseWithError(2, "registration rejected")
		return
	}

	slog.Info("Connected to king", "king_addr", kingAddr)

	tc.mu.Lock()
	tc.connections[kingAddr] = conn
	tc.mu.Unlock()

	if tc.onConnected != nil {
		tc.onConnected()
	}

	go func() {
		for {
			dataStream, err := conn.AcceptStream(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Info("King connection closed", "king_addr", kingAddr)
				tc.mu.Lock()
				delete(tc.connections, kingAddr)
				tc.mu.Unlock()
				return
			}

			go handleDataStream(dataStream, localAddrs)
		}
	}()
}

func handleDataStream(stream quic.Stream, localAddrs map[string]string) {
	headerBuf := make([]byte, 2)
	if _, err := io.ReadFull(stream, headerBuf); err != nil {
		stream.Close()
		return
	}

	serviceIDLen := binary.BigEndian.Uint16(headerBuf)
	serviceIDBuf := make([]byte, serviceIDLen)
	if _, err := io.ReadFull(stream, serviceIDBuf); err != nil {
		stream.Close()
		return
	}

	localAddr, ok := localAddrs[string(serviceIDBuf)]
	if !ok {
		stream.Close()
		return
	}

	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		slog.Error("Failed to dial local service", "local_addr", localAddr, "error", err)
		stream.Close()
		return
	}

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(localConn, stream)
		localConn.Close()
		close(done)
	}()
	_, _ = io.Copy(stream, localConn)
	stream.Close()
	stream.CancelRead(0)
	<-done
}

func (tc *tunnelClient) closeAll() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	for addr, conn := range tc.connections {
		_ = conn.CloseWithError(0, "shutting down")
		delete(tc.connections, addr)
	}
}
