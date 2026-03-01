package council

import (
	"sync"
	"testing"
	"time"

	"github.com/firecow/ratnation/internal/state"
)

func TestRunCleaner_RemovesStaleKing(t *testing.T) {
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Beat: time.Now().UnixMilli() - 11000},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Kings) != 0 {
		t.Fatalf("expected 0 kings, got %d", len(s.Kings))
	}
	if s.Revision != 1 {
		t.Fatalf("expected revision 1, got %d", s.Revision)
	}
}

func TestRunCleaner_KeepsFreshKing(t *testing.T) {
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Beat: time.Now().UnixMilli()},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Kings) != 1 {
		t.Fatalf("expected 1 king, got %d", len(s.Kings))
	}
	if s.Revision != 0 {
		t.Fatalf("expected revision 0, got %d", s.Revision)
	}
}

func TestRunCleaner_RemovesStaleLing(t *testing.T) {
	s := &state.State{
		Lings: []state.StateLing{
			{LingID: "ling-1", Beat: time.Now().UnixMilli() - 11000},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Lings) != 0 {
		t.Fatalf("expected 0 lings, got %d", len(s.Lings))
	}
}

func TestRunCleaner_RemovesOrphanedService(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 5000
	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Host:       &host,
				BindPort:   &bindPort,
				RemotePort: &remotePort,
			},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Services) != 0 {
		t.Fatalf("expected 0 services (orphaned), got %d", len(s.Services))
	}
}

func TestRunCleaner_KeepsServiceWithKingAndLing(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 5000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Beat: time.Now().UnixMilli()},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1", Beat: time.Now().UnixMilli()},
		},
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Host:       &host,
				BindPort:   &bindPort,
				RemotePort: &remotePort,
			},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(s.Services))
	}
}

func TestRunCleaner_KeepsUnprovisionedServiceWithLing(t *testing.T) {
	s := &state.State{
		Lings: []state.StateLing{
			{LingID: "ling-1", Beat: time.Now().UnixMilli()},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1"},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Services) != 1 {
		t.Fatalf("expected 1 service (unprovisioned but ling exists), got %d", len(s.Services))
	}
}

func TestRunCleaner_RemovesUnprovisionedServiceWithoutLing(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1"},
		},
	}
	var mu sync.RWMutex
	hub := newWSHub()

	runCleaner(s, &mu, hub)

	if len(s.Services) != 0 {
		t.Fatalf("expected 0 services (ling missing), got %d", len(s.Services))
	}
}
