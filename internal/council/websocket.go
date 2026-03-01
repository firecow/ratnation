package council

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const broadcastTimeout = 5 * time.Second

// WSHub manages WebSocket connections and broadcasting.
type WSHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

// NewWSHub creates a new WebSocket hub.
func NewWSHub() *WSHub {
	return &WSHub{
		mu:      sync.Mutex{},
		clients: make(map[*websocket.Conn]struct{}),
	}
}

// Broadcast sends a state-changed message to all connected WebSocket clients.
func (h *WSHub) Broadcast(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	broadcastCtx, cancel := context.WithTimeout(ctx, broadcastTimeout)
	defer cancel()

	for conn := range h.clients {
		err := conn.Write(
			broadcastCtx,
			websocket.MessageText,
			[]byte("state-changed"),
		)
		if err != nil {
			slog.Warn("Failed to write to WebSocket client", "error", err)

			_ = conn.CloseNow()
			delete(h.clients, conn)
		}
	}
}

func (h *WSHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[conn] = struct{}{}
}

func (h *WSHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, conn)
}

func (h *WSHub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		_ = conn.Close(websocket.StatusGoingAway, "server shutting down")
		delete(h.clients, conn)
	}
}

func sanitizeRemoteAddr(addr string) string {
	return strings.ReplaceAll(strings.ReplaceAll(addr, "\n", ""), "\r", "")
}

func logWebSocketEvent(ctx context.Context, message string, remoteAddr string) {
	sanitized := sanitizeRemoteAddr(remoteAddr)

	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		message,
		slog.String("remote_addr", sanitized),
	)
}

func (h *WSHub) handleWebSocket(
	writer http.ResponseWriter,
	request *http.Request,
) {
	conn, err := websocket.Accept(writer, request, &websocket.AcceptOptions{
		InsecureSkipVerify:   true,
		Subprotocols:         nil,
		OriginPatterns:       nil,
		CompressionMode:      0,
		CompressionThreshold: 0,
		OnPingReceived:       nil,
		OnPongReceived:       nil,
	})
	if err != nil {
		slog.Error("WebSocket accept failed", "error", err)

		return
	}

	remoteAddr := sanitizeRemoteAddr(request.RemoteAddr)
	requestCtx := request.Context()

	logWebSocketEvent(requestCtx, "WebSocket client connected", remoteAddr)
	h.add(conn)

	writeErr := conn.Write(
		requestCtx,
		websocket.MessageText,
		[]byte("state-changed"),
	)
	if writeErr != nil {
		slog.Warn("Failed to send initial state-changed", "error", writeErr)
	}

	for {
		_, _, readErr := conn.Read(requestCtx)
		if readErr != nil {
			break
		}
	}

	h.remove(conn)
	logWebSocketEvent(requestCtx, "WebSocket client disconnected", remoteAddr)
}
