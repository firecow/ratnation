package king

import (
	"testing"

	"github.com/firecow/burrow/internal/state"
)

// --- parseTunnelArgs ---

func TestParseTunnelArgs_ValidSingle(t *testing.T) {
	configs, err := parseTunnelArgs([]string{"bind_port=2333 ports=5000-5100"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].BindPort != 2333 {
		t.Fatalf("expected bind_port 2333, got %d", configs[0].BindPort)
	}
	if configs[0].Ports != "5000-5100" {
		t.Fatalf("expected ports 5000-5100, got %s", configs[0].Ports)
	}
}

func TestParseTunnelArgs_ValidMultiple(t *testing.T) {
	configs, err := parseTunnelArgs([]string{
		"bind_port=2333 ports=5000-5100",
		"bind_port=2334 ports=6000-6050",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}
	if configs[0].BindPort != 2333 {
		t.Fatalf("expected first bind_port 2333, got %d", configs[0].BindPort)
	}
	if configs[1].BindPort != 2334 {
		t.Fatalf("expected second bind_port 2334, got %d", configs[1].BindPort)
	}
	if configs[1].Ports != "6000-6050" {
		t.Fatalf("expected second ports 6000-6050, got %s", configs[1].Ports)
	}
}

func TestParseTunnelArgs_Empty(t *testing.T) {
	configs, err := parseTunnelArgs(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 0 {
		t.Fatalf("expected 0 configs, got %d", len(configs))
	}
}

func TestParseTunnelArgs_MissingBindPort(t *testing.T) {
	_, err := parseTunnelArgs([]string{"ports=5000-5100"})
	if err == nil {
		t.Fatal("expected error for missing bind_port")
	}
}

func TestParseTunnelArgs_MissingPorts(t *testing.T) {
	_, err := parseTunnelArgs([]string{"bind_port=2333"})
	if err == nil {
		t.Fatal("expected error for missing ports")
	}
}

func TestParseTunnelArgs_InvalidBindPort(t *testing.T) {
	_, err := parseTunnelArgs([]string{"bind_port=abc ports=5000-5100"})
	if err == nil {
		t.Fatal("expected error for invalid bind_port")
	}
}

// --- buildSyncTunnels ---

func TestBuildSyncTunnels_Empty(t *testing.T) {
	result := buildSyncTunnels(nil)
	if len(result) != 0 {
		t.Fatalf("expected 0 sync tunnels, got %d", len(result))
	}
}

func TestBuildSyncTunnels_Single(t *testing.T) {
	result := buildSyncTunnels([]tunnelConfig{
		{BindPort: 2333, Ports: "5000-5100"},
	})
	if len(result) != 1 {
		t.Fatalf("expected 1 sync tunnel, got %d", len(result))
	}
	if result[0].BindPort != 2333 {
		t.Fatalf("expected bind_port 2333, got %d", result[0].BindPort)
	}
	if result[0].Ports != "5000-5100" {
		t.Fatalf("expected ports 5000-5100, got %s", result[0].Ports)
	}
}

func TestBuildSyncTunnels_Multiple(t *testing.T) {
	result := buildSyncTunnels([]tunnelConfig{
		{BindPort: 2333, Ports: "5000-5100"},
		{BindPort: 2334, Ports: "6000-6050"},
	})
	if len(result) != 2 {
		t.Fatalf("expected 2 sync tunnels, got %d", len(result))
	}
	if result[1].BindPort != 2334 {
		t.Fatalf("expected second bind_port 2334, got %d", result[1].BindPort)
	}
}

// --- computeReadyServiceIDs ---

func TestComputeReadyServiceIDs_MatchingService(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	ts := newTunnelServer(2333, nil)
	ts.quicConns["svc-1"] = nil // presence in map is what matters

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1", ShuttingDown: false},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 1 {
		t.Fatalf("expected 1 ready service, got %d", len(ids))
	}
	if ids[0] != "svc-1" {
		t.Fatalf("expected svc-1, got %s", ids[0])
	}
}

func TestComputeReadyServiceIDs_DifferentHost(t *testing.T) {
	host := "9.9.9.9"
	bindPort := 2333
	ts := newTunnelServer(2333, nil)
	ts.quicConns["svc-1"] = nil

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (different host), got %d", len(ids))
	}
}

func TestComputeReadyServiceIDs_NilHostAndBindPort(t *testing.T) {
	ts := newTunnelServer(2333, nil)

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1"},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (nil host/bind_port), got %d", len(ids))
	}
}

func TestComputeReadyServiceIDs_MissingLing(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	ts := newTunnelServer(2333, nil)
	ts.quicConns["svc-1"] = nil

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (missing ling), got %d", len(ids))
	}
}

func TestComputeReadyServiceIDs_ShuttingDownLing(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	ts := newTunnelServer(2333, nil)
	ts.quicConns["svc-1"] = nil

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1", ShuttingDown: true},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (shutting down ling), got %d", len(ids))
	}
}

func TestComputeReadyServiceIDs_NoQUICConnection(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	ts := newTunnelServer(2333, nil)
	// no entry in ts.quicConns

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (no QUIC connection), got %d", len(ids))
	}
}

func TestComputeReadyServiceIDs_NonMatchingTunnel(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 9999
	ts := newTunnelServer(2333, nil)
	ts.quicConns["svc-1"] = nil

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (bind_port mismatch), got %d", len(ids))
	}
}

func TestComputeReadyServiceIDs_MultipleServicesPartialReady(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	ts := newTunnelServer(2333, nil)
	ts.quicConns["svc-1"] = nil
	// svc-2 has no QUIC connection

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Host: &host, BindPort: &bindPort},
			{ServiceID: "svc-2", LingID: "ling-2", Host: &host, BindPort: &bindPort},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
			{LingID: "ling-2"},
		},
	}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 1 {
		t.Fatalf("expected 1 ready service, got %d", len(ids))
	}
	if ids[0] != "svc-1" {
		t.Fatalf("expected svc-1, got %s", ids[0])
	}
}

func TestComputeReadyServiceIDs_EmptyState(t *testing.T) {
	ts := newTunnelServer(2333, nil)
	s := &state.State{}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "5000-5100"}}

	ids := computeReadyServiceIDs(s, tunnels, tunnelConfigs, "1.2.3.4")
	if len(ids) != 0 {
		t.Fatalf("expected 0 ready services (empty state), got %d", len(ids))
	}
}

// --- onStateChanged ---

func TestOnStateChanged_UpdatesServicesMap(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 39701
	ts := newTunnelServer(2333, nil)
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "39701-39710"}}

	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Token:      "tok-1",
				Host:       &host,
				BindPort:   &bindPort,
				RemotePort: &remotePort,
			},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1", ShuttingDown: false},
		},
	}

	onStateChanged(s, tunnels, tunnelConfigs, "1.2.3.4")

	ts.mu.RLock()
	defer ts.mu.RUnlock()
	auth, exists := ts.services["svc-1"]
	if !exists {
		t.Fatal("expected svc-1 in services map")
	}
	if auth.token != "tok-1" {
		t.Fatalf("expected token tok-1, got %s", auth.token)
	}
}

func TestOnStateChanged_ExcludesShuttingDownLing(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 39711
	ts := newTunnelServer(2333, nil)
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "39711-39720"}}

	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Token:      "tok-1",
				Host:       &host,
				BindPort:   &bindPort,
				RemotePort: &remotePort,
			},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1", ShuttingDown: true},
		},
	}

	onStateChanged(s, tunnels, tunnelConfigs, "1.2.3.4")

	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if len(ts.services) != 0 {
		t.Fatalf("expected 0 services (ling shutting down), got %d", len(ts.services))
	}
}

func TestOnStateChanged_ExcludesDifferentHost(t *testing.T) {
	host := "9.9.9.9"
	bindPort := 2333
	remotePort := 39721
	ts := newTunnelServer(2333, nil)
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "39721-39730"}}

	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Token:      "tok-1",
				Host:       &host,
				BindPort:   &bindPort,
				RemotePort: &remotePort,
			},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}

	onStateChanged(s, tunnels, tunnelConfigs, "1.2.3.4")

	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if len(ts.services) != 0 {
		t.Fatalf("expected 0 services (different host), got %d", len(ts.services))
	}
}

func TestOnStateChanged_ExcludesNilFields(t *testing.T) {
	ts := newTunnelServer(2333, nil)
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "39731-39740"}}

	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", LingID: "ling-1", Token: "tok-1"},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}

	onStateChanged(s, tunnels, tunnelConfigs, "1.2.3.4")

	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if len(ts.services) != 0 {
		t.Fatalf("expected 0 services (nil fields), got %d", len(ts.services))
	}
}

func TestOnStateChanged_MultipleTunnels(t *testing.T) {
	host := "1.2.3.4"
	bindPort1 := 2333
	bindPort2 := 2334
	remotePort1 := 39741
	remotePort2 := 39751
	ts1 := newTunnelServer(2333, nil)
	ts2 := newTunnelServer(2334, nil)
	tunnels := map[int]*tunnelServer{2333: ts1, 2334: ts2}
	tunnelConfigs := []tunnelConfig{
		{BindPort: 2333, Ports: "39741-39750"},
		{BindPort: 2334, Ports: "39751-39760"},
	}

	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Token:      "tok-1",
				Host:       &host,
				BindPort:   &bindPort1,
				RemotePort: &remotePort1,
			},
			{
				ServiceID:  "svc-2",
				LingID:     "ling-2",
				Token:      "tok-2",
				Host:       &host,
				BindPort:   &bindPort2,
				RemotePort: &remotePort2,
			},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
			{LingID: "ling-2"},
		},
	}

	onStateChanged(s, tunnels, tunnelConfigs, "1.2.3.4")

	ts1.mu.RLock()
	_, hasSvc1 := ts1.services["svc-1"]
	_, hasSvc2InTs1 := ts1.services["svc-2"]
	ts1.mu.RUnlock()

	ts2.mu.RLock()
	_, hasSvc2 := ts2.services["svc-2"]
	_, hasSvc1InTs2 := ts2.services["svc-1"]
	ts2.mu.RUnlock()

	if !hasSvc1 {
		t.Fatal("expected svc-1 in ts1 services")
	}
	if hasSvc2InTs1 {
		t.Fatal("svc-2 should not be in ts1")
	}
	if !hasSvc2 {
		t.Fatal("expected svc-2 in ts2 services")
	}
	if hasSvc1InTs2 {
		t.Fatal("svc-1 should not be in ts2")
	}
}

func TestOnStateChanged_ClearsServicesOnEmpty(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 39761
	ts := newTunnelServer(2333, nil)
	ts.services["svc-old"] = serviceAuth{token: "old-tok"}
	tunnels := map[int]*tunnelServer{2333: ts}
	tunnelConfigs := []tunnelConfig{{BindPort: 2333, Ports: "39761-39770"}}

	// First call with a valid service
	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				LingID:     "ling-1",
				Token:      "tok-1",
				Host:       &host,
				BindPort:   &bindPort,
				RemotePort: &remotePort,
			},
		},
		Lings: []state.StateLing{
			{LingID: "ling-1"},
		},
	}
	onStateChanged(s, tunnels, tunnelConfigs, "1.2.3.4")

	ts.mu.RLock()
	if len(ts.services) != 1 {
		ts.mu.RUnlock()
		t.Fatalf("expected 1 service after first call, got %d", len(ts.services))
	}
	ts.mu.RUnlock()

	// Second call with empty services replaces the map
	s2 := &state.State{}
	onStateChanged(s2, tunnels, tunnelConfigs, "1.2.3.4")

	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if len(ts.services) != 0 {
		t.Fatalf("expected 0 services after clearing, got %d", len(ts.services))
	}
}

// --- syncer.triggerSync ---

func TestTriggerSync_SendsNotification(t *testing.T) {
	s := &syncer{
		notify: make(chan struct{}, 1),
	}

	s.triggerSync()

	select {
	case <-s.notify:
	default:
		t.Fatal("expected notification on notify channel")
	}
}

func TestTriggerSync_DoesNotBlockWhenFull(t *testing.T) {
	s := &syncer{
		notify: make(chan struct{}, 1),
	}

	// Fill the channel
	s.notify <- struct{}{}

	// Should not block
	s.triggerSync()

	// Drain and verify there's exactly one message
	<-s.notify
	select {
	case <-s.notify:
		t.Fatal("expected only one notification in channel")
	default:
	}
}
