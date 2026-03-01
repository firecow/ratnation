package king

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type syncPayload struct {
	Host            string       `json:"host"`
	ShuttingDown    bool         `json:"shutting_down"`
	Tunnels         []syncTunnel `json:"tunnels"`
	ReadyServiceIDs []string     `json:"ready_service_ids"`
	Location        string       `json:"location"`
	CertPEM         string       `json:"cert_pem"`
}

type syncTunnel struct {
	BindPort int    `json:"bind_port"`
	Ports    string `json:"ports"`
}

type syncer struct {
	councilHost     string
	host            string
	location        string
	certPEM         string
	tunnels         []syncTunnel
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
		Tunnels:         s.tunnels,
		ReadyServiceIDs: s.readyServiceIDs(),
		Location:        s.location,
		CertPEM:         s.certPEM,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal king sync payload", "error", err)
		return
	}

	baseURL, err := url.Parse(s.councilHost)
	if err != nil {
		slog.Error("Failed to parse council host URL", "error", err)
		return
	}
	syncURL := baseURL.JoinPath("/king")

	req := &http.Request{
		Method: http.MethodPut,
		URL:    syncURL,
		Host:   syncURL.Host,
		Header: http.Header{"Content-Type": {"application/json"}},
		Body:   io.NopCloser(bytes.NewReader(body)),
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to sync with council", "error", err)
		return
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to sync with council", "status_code", resp.StatusCode)
	}
}
