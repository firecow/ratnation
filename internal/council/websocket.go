package council

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type wsHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newWSHub() *wsHub {
	return &wsHub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (h *wsHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
}

func (h *wsHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
}

func (h *wsHub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for conn := range h.clients {
		err := conn.Write(ctx, websocket.MessageText, []byte("state-changed"))
		if err != nil {
			slog.Warn("Failed to write to WebSocket client", "error", err)
			_ = conn.CloseNow()
			delete(h.clients, conn)
		}
	}
}

func (h *wsHub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		conn.Close(websocket.StatusGoingAway, "server shutting down")
		delete(h.clients, conn)
	}
}

func (h *wsHub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("WebSocket accept failed", "error", err)
		return
	}

	slog.Info("WebSocket client connected", "remote_addr", r.RemoteAddr)
	h.add(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conn.Write(ctx, websocket.MessageText, []byte("state-changed")); err != nil {
		slog.Warn("Failed to send initial state-changed", "error", err)
	}

	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			break
		}
	}

	h.remove(conn)
	slog.Info("WebSocket client disconnected", "remote_addr", r.RemoteAddr)
}
