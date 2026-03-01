package ling

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
	LingID            string       `json:"ling_id"`
	ShuttingDown      bool         `json:"shutting_down"`
	Tunnels           []syncTunnel `json:"tunnels"`
	ReadyServiceIDs   []string     `json:"ready_service_ids"`
	PreferredLocation string       `json:"preferred_location"`
}

type syncTunnel struct {
	Name string `json:"name"`
}

type lingSyncer struct {
	councilHost       string
	lingID            string
	preferredLocation string
	tunnels           []syncTunnel
	readyServiceIDs   func() []string
	shuttingDown      func() bool
	notify            chan struct{}
}

func (s *lingSyncer) run(ctx context.Context) {
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

func (s *lingSyncer) triggerSync() {
	select {
	case s.notify <- struct{}{}:
	default:
	}
}

func (s *lingSyncer) sync(ctx context.Context) {
	readyIDs := s.readyServiceIDs()
	if readyIDs == nil {
		readyIDs = []string{}
	}

	payload := syncPayload{
		LingID:            s.lingID,
		ShuttingDown:      s.shuttingDown(),
		Tunnels:           s.tunnels,
		ReadyServiceIDs:   readyIDs,
		PreferredLocation: s.preferredLocation,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal ling sync payload", "error", err)
		return
	}

	baseURL, err := url.Parse(s.councilHost)
	if err != nil {
		slog.Error("Failed to parse council host URL", "error", err)
		return
	}
	syncURL := baseURL.JoinPath("/ling")

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
