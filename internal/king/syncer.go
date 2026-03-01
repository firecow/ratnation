package king

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/firecow/burrow/internal/state"
)

const (
	syncTickerInterval = 1 * time.Second
)

// SyncPayload represents the payload sent to the council during sync.
type SyncPayload struct {
	Host            string       `json:"host"`
	ShuttingDown    bool         `json:"shuttingDown"`
	Tunnels         []SyncTunnel `json:"tunnels"`
	ReadyServiceIDs []string     `json:"readyServiceIds"`
	Location        string       `json:"location"`
	CertPEM         string       `json:"certPem"`
}

// SyncTunnel represents a tunnel configuration in the sync payload.
type SyncTunnel struct {
	BindPort int    `json:"bindPort"`
	Ports    string `json:"ports"`
}

// Syncer handles periodic synchronization with the council.
type Syncer struct {
	councilHost    string
	host           string
	location       string
	certPEM        string
	tunnels        []SyncTunnel
	notify         chan struct{}
	mutex          *sync.Mutex
	currentState   *state.State
	isShuttingDown bool
	tunnelServers  map[int]*TunnelServer
	tunnelConfigs  []TunnelConfig
	httpTransport  http.RoundTripper
}

// NewSyncer creates a new Syncer instance.
func NewSyncer(
	councilHost, host, location, certPEM string,
	tunnels []SyncTunnel,
	notify chan struct{},
	mutex *sync.Mutex,
	tunnelServers map[int]*TunnelServer,
	tunnelConfigs []TunnelConfig,
) *Syncer {
	return &Syncer{
		councilHost:    councilHost,
		host:           host,
		location:       location,
		certPEM:        certPEM,
		tunnels:        tunnels,
		notify:         notify,
		mutex:          mutex,
		currentState:   nil,
		isShuttingDown: false,
		tunnelServers:  tunnelServers,
		tunnelConfigs:  tunnelConfigs,
		httpTransport:  http.DefaultTransport,
	}
}

// TriggerSync exposes triggerSync for testing.
func (syncer *Syncer) TriggerSync() {
	syncer.triggerSync()
}

// Notify returns the notify channel for testing.
func (syncer *Syncer) Notify() chan struct{} {
	return syncer.notify
}

func (syncer *Syncer) run(ctx context.Context) {
	syncer.sync(ctx)

	ticker := time.NewTicker(syncTickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			syncer.sync(ctx)
		case <-syncer.notify:
			syncer.sync(ctx)
		}
	}
}

func (syncer *Syncer) triggerSync() {
	select {
	case syncer.notify <- struct{}{}:
	default:
	}
}

func (syncer *Syncer) readyServiceIDs() []string {
	syncer.mutex.Lock()
	defer syncer.mutex.Unlock()

	if syncer.currentState == nil {
		return []string{}
	}

	return ComputeReadyServiceIDs(
		syncer.currentState,
		syncer.tunnelServers,
		syncer.tunnelConfigs,
		syncer.host,
	)
}

func (syncer *Syncer) shuttingDown() bool {
	syncer.mutex.Lock()
	defer syncer.mutex.Unlock()

	return syncer.isShuttingDown
}

func (syncer *Syncer) sync(ctx context.Context) {
	payload := syncer.buildPayload()

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error(
			"Failed to marshal king sync payload",
			"error", err,
		)

		return
	}

	syncer.sendPayload(ctx, body)
}

func (syncer *Syncer) buildPayload() SyncPayload {
	return SyncPayload{
		Host:            syncer.host,
		ShuttingDown:    syncer.shuttingDown(),
		Tunnels:         syncer.tunnels,
		ReadyServiceIDs: syncer.readyServiceIDs(),
		Location:        syncer.location,
		CertPEM:         syncer.certPEM,
	}
}

func (syncer *Syncer) sendPayload(
	ctx context.Context, body []byte,
) {
	syncURL, err := buildSyncURL(syncer.councilHost)
	if err != nil {
		slog.Error(
			"Failed to parse council host URL",
			"error", err,
		)

		return
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		syncURL,
		bytes.NewReader(body),
	)
	if err != nil {
		slog.Error(
			"Failed to create sync request",
			"error", err,
		)

		return
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := syncer.httpTransport.RoundTrip(request)
	if err != nil {
		slog.Error(
			"Failed to sync with council",
			"error", err,
		)

		return
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to sync with council")
	}
}

var errInvalidURLScheme = errors.New(
	"council host must use http or https scheme",
)

func buildSyncURL(
	councilHost string,
) (string, error) {
	baseURL, err := url.Parse(councilHost)
	if err != nil {
		return "", fmt.Errorf(
			"parsing council host URL: %w", err,
		)
	}

	if baseURL.Scheme != "http" &&
		baseURL.Scheme != "https" {
		return "", errInvalidURLScheme
	}

	return baseURL.Scheme + "://" +
		baseURL.Host + "/king", nil
}
