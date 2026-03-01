package king

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type syncPayload struct {
	Host            string        `json:"host"`
	ShuttingDown    bool          `json:"shutting_down"`
	Ratholes        []syncRathole `json:"ratholes"`
	ReadyServiceIDs []string      `json:"ready_service_ids"`
	Location        string        `json:"location"`
	CertPEM         string        `json:"cert_pem"`
}

type syncRathole struct {
	BindPort int    `json:"bind_port"`
	Ports    string `json:"ports"`
}

type syncer struct {
	councilHost     string
	host            string
	location        string
	certPEM         string
	ratholes        []syncRathole
	readyServiceIDs func() []string
	shuttingDown    func() bool
	notify          chan struct{}
}

func (s *syncer) run(ctx context.Context) {
	s.sync(ctx)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sync(ctx)
		case <-s.notify:
			s.sync(ctx)
		}
	}
}

func (s *syncer) triggerSync() {
	select {
	case s.notify <- struct{}{}:
	default:
	}
}

func (s *syncer) sync(ctx context.Context) {
	payload := syncPayload{
		Host:            s.host,
		ShuttingDown:    s.shuttingDown(),
		Ratholes:        s.ratholes,
		ReadyServiceIDs: s.readyServiceIDs(),
		Location:        s.location,
		CertPEM:         s.certPEM,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal king sync payload", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.councilHost+"/king", bytes.NewReader(body))
	if err != nil {
		slog.Error("Failed to create king sync request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to sync with council", "error", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to sync with council", "status_code", resp.StatusCode)
	}
}
