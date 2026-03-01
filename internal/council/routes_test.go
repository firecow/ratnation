package council

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/firecow/burrow/internal/state"
)

func newTestState() *state.State {
	return &state.State{
		Revision: 0,
		Services: []state.StateService{},
		Kings:    []state.StateKing{},
		Lings:    []state.StateLing{},
	}
}

func TestGetState_ReturnsJSON(t *testing.T) {
	s := newTestState()
	var mu sync.RWMutex
	handler := handleGetState(s, &mu)

	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result state.State
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.Revision != 0 {
		t.Fatalf("expected revision 0, got %d", result.Revision)
	}
}

func TestPutKing_CreatesNewKing(t *testing.T) {
	s := newTestState()
	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutKing(s, &mu, hub)

	payload := putKingRequest{
		Host:            "1.2.3.4",
		ShuttingDown:    false,
		Tunnels:         []putKingTunnel{{BindPort: 2333, Ports: "5000-5001"}},
		ReadyServiceIDs: []string{},
		Location:        "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/king", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if len(s.Kings) != 1 {
		t.Fatalf("expected 1 king, got %d", len(s.Kings))
	}
	if s.Kings[0].Host != "1.2.3.4" {
		t.Fatalf("expected host 1.2.3.4, got %s", s.Kings[0].Host)
	}
	if s.Kings[0].Location != "CPH" {
		t.Fatalf("expected location CPH, got %s", s.Kings[0].Location)
	}
	if s.Revision != 1 {
		t.Fatalf("expected revision 1, got %d", s.Revision)
	}
}

func TestPutKing_UpdatesExistingKing(t *testing.T) {
	s := newTestState()
	s.Kings = append(s.Kings, state.StateKing{
		BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", Location: "CPH", Beat: 1000,
	})

	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutKing(s, &mu, hub)

	payload := putKingRequest{
		Host:            "1.2.3.4",
		ShuttingDown:    false,
		Tunnels:         []putKingTunnel{{BindPort: 2333, Ports: "5000-5001"}},
		ReadyServiceIDs: []string{},
		Location:        "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/king", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if len(s.Kings) != 1 {
		t.Fatalf("expected 1 king (updated), got %d", len(s.Kings))
	}
	if s.Kings[0].Beat == 1000 {
		t.Fatalf("expected beat to be updated")
	}
}

func TestPutKing_SetsReadyServiceIDs(t *testing.T) {
	s := newTestState()
	s.Kings = append(s.Kings, state.StateKing{
		BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", Location: "CPH",
	})
	s.Services = append(s.Services, state.StateService{
		ServiceID: "svc-1", Name: "alpha", KingReady: false,
	})

	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutKing(s, &mu, hub)

	payload := putKingRequest{
		Host:            "1.2.3.4",
		ShuttingDown:    false,
		Tunnels:         []putKingTunnel{{BindPort: 2333, Ports: "5000-5001"}},
		ReadyServiceIDs: []string{"svc-1"},
		Location:        "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/king", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if !s.Services[0].KingReady {
		t.Fatalf("expected king_ready to be true")
	}
}

func TestPutLing_CreatesNewLingAndService(t *testing.T) {
	s := newTestState()
	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutLing(s, &mu, hub)

	payload := putLingRequest{
		LingID:            "ling-1",
		ShuttingDown:      false,
		Tunnels:           []putLingTunnel{{Name: "alpha"}},
		ReadyServiceIDs:   []string{},
		PreferredLocation: "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/ling", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if len(s.Lings) != 1 {
		t.Fatalf("expected 1 ling, got %d", len(s.Lings))
	}
	if s.Lings[0].LingID != "ling-1" {
		t.Fatalf("expected ling_id ling-1, got %s", s.Lings[0].LingID)
	}

	if len(s.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(s.Services))
	}
	if s.Services[0].Name != "alpha" {
		t.Fatalf("expected service name alpha, got %s", s.Services[0].Name)
	}
	if s.Services[0].LingID != "ling-1" {
		t.Fatalf("expected ling_id ling-1, got %s", s.Services[0].LingID)
	}
	if s.Services[0].Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if s.Services[0].ServiceID == "" {
		t.Fatalf("expected non-empty service_id")
	}
}

func TestPutLing_DoesNotDuplicateService(t *testing.T) {
	s := newTestState()
	s.Lings = append(s.Lings, state.StateLing{LingID: "ling-1", Beat: 1000})
	s.Services = append(s.Services, state.StateService{
		ServiceID: "svc-1", Name: "alpha", LingID: "ling-1",
	})

	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutLing(s, &mu, hub)

	payload := putLingRequest{
		LingID:            "ling-1",
		ShuttingDown:      false,
		Tunnels:           []putLingTunnel{{Name: "alpha"}},
		ReadyServiceIDs:   []string{},
		PreferredLocation: "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/ling", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if len(s.Services) != 1 {
		t.Fatalf("expected 1 service (no duplicate), got %d", len(s.Services))
	}
}

func TestPutLing_SetsReadyServiceIDs(t *testing.T) {
	s := newTestState()
	s.Services = append(s.Services, state.StateService{
		ServiceID: "svc-1", Name: "alpha", LingID: "ling-1", LingReady: false,
	})

	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutLing(s, &mu, hub)

	payload := putLingRequest{
		LingID:            "ling-1",
		ShuttingDown:      false,
		Tunnels:           []putLingTunnel{{Name: "alpha"}},
		ReadyServiceIDs:   []string{"svc-1"},
		PreferredLocation: "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/ling", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if !s.Services[0].LingReady {
		t.Fatalf("expected ling_ready to be true")
	}
}

func TestPutKing_MissingHost(t *testing.T) {
	s := newTestState()
	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutKing(s, &mu, hub)

	payload := putKingRequest{
		Host:            "",
		Tunnels:         []putKingTunnel{{BindPort: 2333, Ports: "5000-5001"}},
		ReadyServiceIDs: []string{},
		Location:        "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/king", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing host, got %d", w.Result().StatusCode)
	}
}

func TestPutLing_MissingLingID(t *testing.T) {
	s := newTestState()
	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutLing(s, &mu, hub)

	payload := putLingRequest{
		LingID:            "",
		Tunnels:           []putLingTunnel{{Name: "alpha"}},
		ReadyServiceIDs:   []string{},
		PreferredLocation: "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/ling", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing ling_id, got %d", w.Result().StatusCode)
	}
}

func TestPutKing_ShuttingDownUpdatesState(t *testing.T) {
	s := newTestState()
	s.Kings = append(s.Kings, state.StateKing{
		BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", Location: "CPH", Beat: 1000,
	})

	var mu sync.RWMutex
	hub := newWSHub()
	handler := handlePutKing(s, &mu, hub)

	payload := putKingRequest{
		Host:            "1.2.3.4",
		ShuttingDown:    true,
		Tunnels:         []putKingTunnel{{BindPort: 2333, Ports: "5000-5001"}},
		ReadyServiceIDs: []string{},
		Location:        "CPH",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/king", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if !s.Kings[0].ShuttingDown {
		t.Fatalf("expected shutting_down to be true")
	}
}

func TestFullProvisioningFlow(t *testing.T) {
	s := newTestState()
	var mu sync.RWMutex
	hub := newWSHub()

	// 1. King registers
	kingHandler := handlePutKing(s, &mu, hub)
	kingPayload := putKingRequest{
		Host:            "1.2.3.4",
		Tunnels:         []putKingTunnel{{BindPort: 2333, Ports: "5000-5001"}},
		ReadyServiceIDs: []string{},
		Location:        "CPH",
	}
	body, _ := json.Marshal(kingPayload)
	req := httptest.NewRequest(http.MethodPut, "/king", bytes.NewReader(body))
	w := httptest.NewRecorder()
	kingHandler(w, req)

	// 2. Ling registers with a service
	lingHandler := handlePutLing(s, &mu, hub)
	lingPayload := putLingRequest{
		LingID:            "ling-1",
		Tunnels:           []putLingTunnel{{Name: "alpha"}},
		ReadyServiceIDs:   []string{},
		PreferredLocation: "CPH",
	}
	body, _ = json.Marshal(lingPayload)
	req = httptest.NewRequest(http.MethodPut, "/ling", bytes.NewReader(body))
	w = httptest.NewRecorder()
	lingHandler(w, req)

	// Service should be provisioned (king with available ports exists)
	if len(s.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(s.Services))
	}
	svc := s.Services[0]
	if svc.Host == nil || *svc.Host != "1.2.3.4" {
		t.Fatalf("expected provisioned to host 1.2.3.4")
	}
	if svc.RemotePort == nil || *svc.RemotePort != 5000 {
		t.Fatalf("expected remote_port 5000, got %v", svc.RemotePort)
	}
}
