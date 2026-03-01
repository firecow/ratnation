package state

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Watcher struct {
	councilHost string
	state       *State
	mu          sync.Mutex
	onChange    func(*State)
}

func NewWatcher(councilHost string, onChange func(*State)) *Watcher {
	return &Watcher{
		councilHost: councilHost,
		onChange:    onChange,
	}
}

func (w *Watcher) Run(ctx context.Context) {
	go w.listenWebSocket(ctx)

	// Poll immediately on start
	w.fetchState(ctx)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.fetchState(ctx)
		}
	}
}

func (w *Watcher) WaitForState(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.mu.Lock()
			hasState := w.state != nil
			w.mu.Unlock()
			if hasState {
				return nil
			}
		}
	}
}

func (w *Watcher) fetchState(ctx context.Context) {
	url := w.councilHost + "/state"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("Failed to create state request", "error", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to fetch state from council", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to fetch state from council", "status_code", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read state response", "error", err)
		return
	}

	var newState State
	if err := json.Unmarshal(body, &newState); err != nil {
		slog.Error("Failed to parse state response", "error", err)
		return
	}

	w.mu.Lock()
	if w.state == nil || w.state.Revision != newState.Revision {
		w.state = &newState
		w.mu.Unlock()
		w.onChange(&newState)
		return
	}
	w.mu.Unlock()
}

func (w *Watcher) listenWebSocket(ctx context.Context) {
	wsURL := httpToWS(w.councilHost) + "/ws"

	for {
		if ctx.Err() != nil {
			return
		}

		conn, resp, err := websocket.Dial(ctx, wsURL, nil)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if err != nil {
			slog.Error("WebSocket connection failed", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		slog.Info("WebSocket connected")

		for {
			_, message, err := conn.Read(ctx)
			if err != nil {
				slog.Info("WebSocket disconnected", "error", err)
				break
			}
			if string(message) == "state-changed" {
				slog.Info("State changed event received, force fetching")
				w.fetchState(ctx)
			}
		}

		_ = conn.CloseNow()
	}
}

func httpToWS(url string) string {
	if len(url) > 7 && url[:8] == "https://" {
		return "wss://" + url[8:]
	}
	if len(url) > 6 && url[:7] == "http://" {
		return "ws://" + url[7:]
	}
	return "ws://" + url
}
