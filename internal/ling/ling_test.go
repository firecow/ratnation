package ling

import (
	"context"
	"testing"

	"github.com/firecow/burrow/internal/state"
)

func TestParseTunnelArgs_Valid(t *testing.T) {
	args := []string{"name=myservice local_addr=127.0.0.1:8080"}
	configs, err := parseTunnelArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Name != "myservice" {
		t.Errorf("expected name=myservice, got %s", configs[0].Name)
	}
	if configs[0].LocalAddr != "127.0.0.1:8080" {
		t.Errorf("expected local_addr=127.0.0.1:8080, got %s", configs[0].LocalAddr)
	}
}

func TestParseTunnelArgs_Multiple(t *testing.T) {
	args := []string{
		"name=svc1 local_addr=127.0.0.1:8080",
		"name=svc2 local_addr=127.0.0.1:9090",
	}
	configs, err := parseTunnelArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}
	if configs[0].Name != "svc1" || configs[1].Name != "svc2" {
		t.Errorf("unexpected names: %s, %s", configs[0].Name, configs[1].Name)
	}
}

func TestParseTunnelArgs_MissingName(t *testing.T) {
	args := []string{"local_addr=127.0.0.1:8080"}
	_, err := parseTunnelArgs(args)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	expected := "--tunnel must have 'name' field"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestParseTunnelArgs_MissingLocalAddr(t *testing.T) {
	args := []string{"name=myservice"}
	_, err := parseTunnelArgs(args)
	if err == nil {
		t.Fatal("expected error for missing local_addr")
	}
	expected := "--tunnel must have 'local_addr' field"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestParseTunnelArgs_Empty(t *testing.T) {
	configs, err := parseTunnelArgs(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configs != nil {
		t.Errorf("expected nil configs for nil args, got %v", configs)
	}
}

func TestParseProxyArgs_Valid(t *testing.T) {
	args := []string{"name=myproxy bind_port=3306"}
	configs, err := parseProxyArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Name != "myproxy" {
		t.Errorf("expected name=myproxy, got %s", configs[0].Name)
	}
	if configs[0].BindPort != 3306 {
		t.Errorf("expected bind_port=3306, got %d", configs[0].BindPort)
	}
}

func TestParseProxyArgs_MissingName(t *testing.T) {
	args := []string{"bind_port=3306"}
	_, err := parseProxyArgs(args)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	expected := "--proxy must have 'name' field"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestParseProxyArgs_MissingBindPort(t *testing.T) {
	args := []string{"name=myproxy"}
	_, err := parseProxyArgs(args)
	if err == nil {
		t.Fatal("expected error for missing bind_port")
	}
	expected := "--proxy must have 'bind_port' field"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestParseProxyArgs_InvalidBindPort(t *testing.T) {
	args := []string{"name=myproxy bind_port=abc"}
	_, err := parseProxyArgs(args)
	if err == nil {
		t.Fatal("expected error for invalid bind_port")
	}
	expected := "invalid bind_port: abc"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestParseProxyArgs_Empty(t *testing.T) {
	configs, err := parseProxyArgs(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configs != nil {
		t.Errorf("expected nil configs for nil args, got %v", configs)
	}
}

func TestComputeReadyServiceIDs_MatchingLing(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID: "svc-1",
				Name:      "web",
				LingID:    "ling-a",
				KingReady: true,
			},
		},
	}
	tunnelMap := map[string]string{"web": "127.0.0.1:8080"}
	ids := computeReadyServiceIDs(s, "ling-a", tunnelMap)
	if len(ids) != 1 || ids[0] != "svc-1" {
		t.Errorf("expected [svc-1], got %v", ids)
	}
}

func TestComputeReadyServiceIDs_NonMatchingLing(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID: "svc-1",
				Name:      "web",
				LingID:    "ling-b",
				KingReady: true,
			},
		},
	}
	tunnelMap := map[string]string{"web": "127.0.0.1:8080"}
	ids := computeReadyServiceIDs(s, "ling-a", tunnelMap)
	if len(ids) != 0 {
		t.Errorf("expected empty, got %v", ids)
	}
}

func TestComputeReadyServiceIDs_MissingTunnelEntry(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID: "svc-1",
				Name:      "db",
				LingID:    "ling-a",
				KingReady: true,
			},
		},
	}
	tunnelMap := map[string]string{"web": "127.0.0.1:8080"}
	ids := computeReadyServiceIDs(s, "ling-a", tunnelMap)
	if len(ids) != 0 {
		t.Errorf("expected empty, got %v", ids)
	}
}

func TestComputeReadyServiceIDs_KingNotReady(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{
				ServiceID: "svc-1",
				Name:      "web",
				LingID:    "ling-a",
				KingReady: false,
			},
		},
	}
	tunnelMap := map[string]string{"web": "127.0.0.1:8080"}
	ids := computeReadyServiceIDs(s, "ling-a", tunnelMap)
	if len(ids) != 0 {
		t.Errorf("expected empty, got %v", ids)
	}
}

func TestComputeReadyServiceIDs_MultipleServices(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "web", LingID: "ling-a", KingReady: true},
			{ServiceID: "svc-2", Name: "api", LingID: "ling-a", KingReady: true},
			{ServiceID: "svc-3", Name: "web", LingID: "ling-b", KingReady: true},
			{ServiceID: "svc-4", Name: "db", LingID: "ling-a", KingReady: true},
			{ServiceID: "svc-5", Name: "web", LingID: "ling-a", KingReady: false},
		},
	}
	tunnelMap := map[string]string{
		"web": "127.0.0.1:8080",
		"api": "127.0.0.1:9090",
	}
	ids := computeReadyServiceIDs(s, "ling-a", tunnelMap)
	if len(ids) != 2 {
		t.Fatalf("expected 2 ready IDs, got %d: %v", len(ids), ids)
	}
	idSet := map[string]bool{}
	for _, id := range ids {
		idSet[id] = true
	}
	if !idSet["svc-1"] || !idSet["svc-2"] {
		t.Errorf("expected svc-1 and svc-2, got %v", ids)
	}
}

func TestOnStateChanged_UpdatesProxyTargets(t *testing.T) {
	kingHost := "10.0.0.1"
	kingBindPort := 5000
	remotePort := 12345

	s := &state.State{
		Kings: []state.StateKing{
			{Host: kingHost, BindPort: kingBindPort, ShuttingDown: false},
		},
		Lings: []state.StateLing{
			{LingID: "ling-a", ShuttingDown: false},
		},
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				Name:       "myproxy",
				LingID:     "ling-a",
				LingReady:  true,
				KingReady:  true,
				Host:       &kingHost,
				BindPort:   &kingBindPort,
				RemotePort: &remotePort,
			},
		},
	}

	proxy := newTCPProxy("myproxy", 0)
	tcpProxies := map[string]*tcpProxy{"myproxy": proxy}
	tc := newTunnelClient()

	onStateChanged(context.Background(), s, "other-ling", map[string]string{}, tc, tcpProxies)

	proxy.mu.RLock()
	targets := proxy.targets
	proxy.mu.RUnlock()

	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].host != kingHost {
		t.Errorf("expected host=%s, got %s", kingHost, targets[0].host)
	}
	if targets[0].remotePort != remotePort {
		t.Errorf("expected remotePort=%d, got %d", remotePort, targets[0].remotePort)
	}
}

func TestOnStateChanged_ExcludesShuttingDownLing(t *testing.T) {
	kingHost := "10.0.0.1"
	kingBindPort := 5000
	remotePort := 12345

	s := &state.State{
		Kings: []state.StateKing{
			{Host: kingHost, BindPort: kingBindPort, ShuttingDown: false},
		},
		Lings: []state.StateLing{
			{LingID: "ling-a", ShuttingDown: true},
		},
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				Name:       "myproxy",
				LingID:     "ling-a",
				LingReady:  true,
				KingReady:  true,
				Host:       &kingHost,
				BindPort:   &kingBindPort,
				RemotePort: &remotePort,
			},
		},
	}

	proxy := newTCPProxy("myproxy", 0)
	tcpProxies := map[string]*tcpProxy{"myproxy": proxy}
	tc := newTunnelClient()

	onStateChanged(context.Background(), s, "other-ling", map[string]string{}, tc, tcpProxies)

	proxy.mu.RLock()
	targets := proxy.targets
	proxy.mu.RUnlock()

	if len(targets) != 0 {
		t.Errorf("expected 0 targets (ling shutting down), got %d", len(targets))
	}
}

func TestOnStateChanged_ExcludesShuttingDownKing(t *testing.T) {
	kingHost := "10.0.0.1"
	kingBindPort := 5000
	remotePort := 12345

	s := &state.State{
		Kings: []state.StateKing{
			{Host: kingHost, BindPort: kingBindPort, ShuttingDown: true},
		},
		Lings: []state.StateLing{
			{LingID: "ling-a", ShuttingDown: false},
		},
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				Name:       "myproxy",
				LingID:     "ling-a",
				LingReady:  true,
				KingReady:  true,
				Host:       &kingHost,
				BindPort:   &kingBindPort,
				RemotePort: &remotePort,
			},
		},
	}

	proxy := newTCPProxy("myproxy", 0)
	tcpProxies := map[string]*tcpProxy{"myproxy": proxy}
	tc := newTunnelClient()

	onStateChanged(context.Background(), s, "other-ling", map[string]string{}, tc, tcpProxies)

	proxy.mu.RLock()
	targets := proxy.targets
	proxy.mu.RUnlock()

	if len(targets) != 0 {
		t.Errorf("expected 0 targets (king shutting down), got %d", len(targets))
	}
}

func TestOnStateChanged_ExcludesNotReady(t *testing.T) {
	kingHost := "10.0.0.1"
	kingBindPort := 5000
	remotePort := 12345

	s := &state.State{
		Kings: []state.StateKing{
			{Host: kingHost, BindPort: kingBindPort, ShuttingDown: false},
		},
		Lings: []state.StateLing{
			{LingID: "ling-a", ShuttingDown: false},
		},
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				Name:       "myproxy",
				LingID:     "ling-a",
				LingReady:  false,
				KingReady:  true,
				Host:       &kingHost,
				BindPort:   &kingBindPort,
				RemotePort: &remotePort,
			},
		},
	}

	proxy := newTCPProxy("myproxy", 0)
	tcpProxies := map[string]*tcpProxy{"myproxy": proxy}
	tc := newTunnelClient()

	onStateChanged(context.Background(), s, "other-ling", map[string]string{}, tc, tcpProxies)

	proxy.mu.RLock()
	targets := proxy.targets
	proxy.mu.RUnlock()

	if len(targets) != 0 {
		t.Errorf("expected 0 targets (ling not ready), got %d", len(targets))
	}
}

func TestOnStateChanged_ExcludesMissingHostOrPort(t *testing.T) {
	kingHost := "10.0.0.1"
	kingBindPort := 5000

	s := &state.State{
		Kings: []state.StateKing{
			{Host: kingHost, BindPort: kingBindPort, ShuttingDown: false},
		},
		Lings: []state.StateLing{
			{LingID: "ling-a", ShuttingDown: false},
		},
		Services: []state.StateService{
			{
				ServiceID:  "svc-1",
				Name:       "myproxy",
				LingID:     "ling-a",
				LingReady:  true,
				KingReady:  true,
				Host:       nil,
				RemotePort: nil,
			},
		},
	}

	proxy := newTCPProxy("myproxy", 0)
	tcpProxies := map[string]*tcpProxy{"myproxy": proxy}
	tc := newTunnelClient()

	onStateChanged(context.Background(), s, "other-ling", map[string]string{}, tc, tcpProxies)

	proxy.mu.RLock()
	targets := proxy.targets
	proxy.mu.RUnlock()

	if len(targets) != 0 {
		t.Errorf("expected 0 targets (nil host/port), got %d", len(targets))
	}
}
