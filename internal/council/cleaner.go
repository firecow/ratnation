package council

import (
	"log/slog"
	"sync"
	"time"

	"github.com/firecow/ratnation/internal/state"
)

func startCleaner(s *state.State, mu *sync.RWMutex, hub *wsHub, stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			runCleaner(s, mu, hub)
		}
	}
}

func runCleaner(s *state.State, mu *sync.RWMutex, hub *wsHub) {
	mu.Lock()

	stateChanged := false
	now := time.Now().UnixMilli()
	staleThreshold := now - 10000

	// Remove stale kings
	kings := s.Kings[:0]
	for _, king := range s.Kings {
		if king.Beat <= staleThreshold {
			slog.Info("Removing stale king", "host", king.Host, "bind_port", king.BindPort)
			stateChanged = true
			continue
		}
		kings = append(kings, king)
	}
	s.Kings = kings

	// Remove stale lings
	lings := s.Lings[:0]
	for _, ling := range s.Lings {
		if ling.Beat <= staleThreshold {
			slog.Info("Removing stale ling", "ling_id", ling.LingID)
			stateChanged = true
			continue
		}
		lings = append(lings, ling)
	}
	s.Lings = lings

	// Remove services whose king or ling is missing
	services := s.Services[:0]
	for _, svc := range s.Services {
		kingFound := false
		for _, king := range s.Kings {
			if svc.Host != nil && svc.BindPort != nil && king.Host == *svc.Host && king.BindPort == *svc.BindPort {
				kingFound = true
				break
			}
		}

		lingFound := false
		for _, ling := range s.Lings {
			if ling.LingID == svc.LingID {
				lingFound = true
				break
			}
		}

		// Unprovisioned services (no king assigned) only need their ling
		if svc.BindPort == nil {
			if lingFound {
				services = append(services, svc)
			} else {
				slog.Info("Removing orphaned service (ling missing)", "name", svc.Name, "service_id", svc.ServiceID)
				stateChanged = true
			}
			continue
		}

		if kingFound && lingFound {
			services = append(services, svc)
		} else {
			slog.Info("Removing orphaned service", "name", svc.Name, "service_id", svc.ServiceID, "king_found", kingFound, "ling_found", lingFound)
			stateChanged = true
		}
	}
	s.Services = services

	if stateChanged {
		s.Revision++
		mu.Unlock()
		hub.broadcast()
		return
	}

	mu.Unlock()
}
