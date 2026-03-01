package council

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/firecow/ratnation/internal/state"
	"github.com/google/uuid"
)

type putKingRequest struct {
	Host            string           `json:"host"`
	ShuttingDown    bool             `json:"shutting_down"`
	Ratholes        []putKingRathole `json:"ratholes"`
	ReadyServiceIDs []string         `json:"ready_service_ids"`
	Location        string           `json:"location"`
	CertPEM         string           `json:"cert_pem"`
}

type putKingRathole struct {
	BindPort int    `json:"bind_port"`
	Ports    string `json:"ports"`
}

type putLingRequest struct {
	LingID            string           `json:"ling_id"`
	ShuttingDown      bool             `json:"shutting_down"`
	Ratholes          []putLingRathole `json:"ratholes"`
	ReadyServiceIDs   []string         `json:"ready_service_ids"`
	PreferredLocation string           `json:"preferred_location"`
}

type putLingRathole struct {
	Name string `json:"name"`
}

func handleGetState(s *state.State, mu *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(s)
	}
}

func handlePutKing(s *state.State, mu *sync.RWMutex, hub *wsHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var req putKingRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if req.Host == "" {
			http.Error(w, "host field cannot be null or undefined", http.StatusBadRequest)
			return
		}
		if req.Location == "" {
			http.Error(w, "location field cannot be null or undefined", http.StatusBadRequest)
			return
		}

		mu.Lock()

		for _, serviceID := range req.ReadyServiceIDs {
			svc := findService(s, serviceID)
			if svc == nil {
				mu.Unlock()
				http.Error(w, fmt.Sprintf("%s cannot be found in state.services", serviceID), http.StatusBadRequest)
				return
			}
			if !svc.KingReady {
				svc.KingReady = true
				s.Revision++
				provision(s)
				mu.Unlock()
				hub.broadcast()
				mu.Lock()
			}
		}

		now := time.Now().UnixMilli()

		for _, rathole := range req.Ratholes {
			existingKing := findKing(s, rathole.Ports, req.Host)
			if existingKing != nil {
				changed := existingKing.ShuttingDown != req.ShuttingDown || existingKing.CertPEM != req.CertPEM
				existingKing.ShuttingDown = req.ShuttingDown
				existingKing.CertPEM = req.CertPEM
				if changed {
					s.Revision++
					provision(s)
					mu.Unlock()
					hub.broadcast()
					mu.Lock()
				}
				existingKing.Beat = now
				continue
			}

			s.Kings = append(s.Kings, state.StateKing{
				BindPort:     rathole.BindPort,
				Ports:        rathole.Ports,
				Host:         req.Host,
				Location:     req.Location,
				Beat:         now,
				ShuttingDown: false,
				CertPEM:      req.CertPEM,
			})
			s.Revision++
			provision(s)
			mu.Unlock()
			hub.broadcast()
			mu.Lock()
		}

		mu.Unlock()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "ok")
	}
}

func handlePutLing(s *state.State, mu *sync.RWMutex, hub *wsHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var req putLingRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if req.LingID == "" {
			http.Error(w, "ling_id field cannot be null or undefined", http.StatusBadRequest)
			return
		}
		if req.PreferredLocation == "" {
			http.Error(w, "preferred_location field cannot be null or undefined", http.StatusBadRequest)
			return
		}

		mu.Lock()

		for _, serviceID := range req.ReadyServiceIDs {
			svc := findService(s, serviceID)
			if svc == nil {
				mu.Unlock()
				http.Error(w, "service is undefined or null", http.StatusBadRequest)
				return
			}
			if !svc.LingReady {
				svc.LingReady = true
				s.Revision++
				provision(s)
				mu.Unlock()
				hub.broadcast()
				mu.Lock()
			}
		}

		now := time.Now().UnixMilli()

		for _, rathole := range req.Ratholes {
			existingLing := findLing(s, req.LingID)
			if existingLing == nil {
				s.Lings = append(s.Lings, state.StateLing{
					LingID:       req.LingID,
					Beat:         now,
					ShuttingDown: req.ShuttingDown,
				})
				existingLing = &s.Lings[len(s.Lings)-1]
			}
			existingLing.Beat = now

			if existingLing.ShuttingDown != req.ShuttingDown {
				existingLing.ShuttingDown = req.ShuttingDown
				s.Revision++
				provision(s)
				mu.Unlock()
				hub.broadcast()
				mu.Lock()
			}

			existingService := findServiceByNameAndLing(s, rathole.Name, req.LingID)
			if existingService != nil {
				mu.Unlock()
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				fmt.Fprint(w, "ok")
				return
			}

			token := generateToken()
			s.Services = append(s.Services, state.StateService{
				ServiceID:         uuid.New().String(),
				Name:              rathole.Name,
				Token:             token,
				PreferredLocation: req.PreferredLocation,
				LingID:            req.LingID,
				LingReady:         false,
				KingReady:         false,
				Host:              nil,
				BindPort:          nil,
				RemotePort:        nil,
			})
			s.Revision++
			provision(s)
			mu.Unlock()
			hub.broadcast()
			mu.Lock()
		}

		mu.Unlock()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "ok")
	}
}

func findService(s *state.State, serviceID string) *state.StateService {
	for i := range s.Services {
		if s.Services[i].ServiceID == serviceID {
			return &s.Services[i]
		}
	}
	return nil
}

func findServiceByNameAndLing(s *state.State, name, lingID string) *state.StateService {
	for i := range s.Services {
		if s.Services[i].Name == name && s.Services[i].LingID == lingID {
			return &s.Services[i]
		}
	}
	return nil
}

func findKing(s *state.State, ports, host string) *state.StateKing {
	for i := range s.Kings {
		if s.Kings[i].Ports == ports && s.Kings[i].Host == host {
			return &s.Kings[i]
		}
	}
	return nil
}

func findLing(s *state.State, lingID string) *state.StateLing {
	for i := range s.Lings {
		if s.Lings[i].LingID == lingID {
			return &s.Lings[i]
		}
	}
	return nil
}

func generateToken() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		slog.Error("Failed to generate random token", "error", err)
	}
	return hex.EncodeToString(b)
}
