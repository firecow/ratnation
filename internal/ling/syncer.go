package ling

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type syncPayload struct {
	LingID            string        `json:"ling_id"`
	ShuttingDown      bool          `json:"shutting_down"`
	Ratholes          []syncRathole `json:"ratholes"`
	ReadyServiceIDs   []string      `json:"ready_service_ids"`
	PreferredLocation string        `json:"preferred_location"`
}

type syncRathole struct {
	Name string `json:"name"`
}

type lingSyncer struct {
	councilHost       string
	lingID            string
	preferredLocation string
	ratholes          []syncRathole
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
		Ratholes:          s.ratholes,
		ReadyServiceIDs:   readyIDs,
		PreferredLocation: s.preferredLocation,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal ling sync payload", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.councilHost+"/ling", bytes.NewReader(body))
	if err != nil {
		slog.Error("Failed to create ling sync request", "error", err)
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
