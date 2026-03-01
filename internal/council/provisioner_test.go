package council

import (
	"testing"

	"github.com/firecow/burrow/internal/state"
)

func TestAvailableKingPorts_NoKings(t *testing.T) {
	s := &state.State{}
	result := availableKingPorts(s)
	if len(result) != 0 {
		t.Fatalf("expected 0 available king ports, got %d", len(result))
	}
}

func TestAvailableKingPorts_ShuttingDownExcluded(t *testing.T) {
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", ShuttingDown: true},
		},
	}
	result := availableKingPorts(s)
	if len(result) != 0 {
		t.Fatalf("expected 0 (shutting down king excluded), got %d", len(result))
	}
}

func TestAvailableKingPorts_AllPortsFree(t *testing.T) {
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5002"},
		},
	}
	result := availableKingPorts(s)
	if len(result) != 1 {
		t.Fatalf("expected 1 king, got %d", len(result))
	}
	if len(result[0].ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(result[0].ports))
	}
	if result[0].ports[0] != 5000 || result[0].ports[1] != 5001 || result[0].ports[2] != 5002 {
		t.Fatalf("unexpected ports: %v", result[0].ports)
	}
}

func TestAvailableKingPorts_SomePortsUsed(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 5000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: bindPort, Host: host, Ports: "5000-5002"},
		},
		Services: []state.StateService{
			{Host: &host, BindPort: &bindPort, RemotePort: &remotePort},
		},
	}
	result := availableKingPorts(s)
	if len(result) != 1 {
		t.Fatalf("expected 1 king, got %d", len(result))
	}
	if len(result[0].ports) != 2 {
		t.Fatalf("expected 2 free ports, got %d", len(result[0].ports))
	}
	if result[0].ports[0] != 5001 || result[0].ports[1] != 5002 {
		t.Fatalf("unexpected ports: %v", result[0].ports)
	}
}

func TestAvailableKingPorts_AllPortsUsed(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	port5000 := 5000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: bindPort, Host: host, Ports: "5000-5000"},
		},
		Services: []state.StateService{
			{Host: &host, BindPort: &bindPort, RemotePort: &port5000},
		},
	}
	result := availableKingPorts(s)
	if len(result) != 0 {
		t.Fatalf("expected 0 (all ports used), got %d", len(result))
	}
}

func TestProvisionService_AssignsFirstAvailable(t *testing.T) {
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001"},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha"},
		},
	}

	provisionService(s, &s.Services[0])

	svc := s.Services[0]
	if svc.Host == nil || *svc.Host != "1.2.3.4" {
		t.Fatalf("expected host 1.2.3.4, got %v", svc.Host)
	}
	if svc.BindPort == nil || *svc.BindPort != 2333 {
		t.Fatalf("expected bind_port 2333, got %v", svc.BindPort)
	}
	if svc.RemotePort == nil || *svc.RemotePort != 5000 {
		t.Fatalf("expected remote_port 5000, got %v", svc.RemotePort)
	}
	if s.Revision != 1 {
		t.Fatalf("expected revision 1, got %d", s.Revision)
	}
}

func TestProvision_MultipleUnprovisioned(t *testing.T) {
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001"},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha"},
			{ServiceID: "svc-2", Name: "beta"},
		},
	}

	provision(s)

	if s.Services[0].RemotePort == nil || *s.Services[0].RemotePort != 5000 {
		t.Fatalf("svc-1: expected remote_port 5000, got %v", s.Services[0].RemotePort)
	}
	if s.Services[1].RemotePort == nil || *s.Services[1].RemotePort != 5001 {
		t.Fatalf("svc-2: expected remote_port 5001, got %v", s.Services[1].RemotePort)
	}
}

func TestProvision_NoKingsNoAssignment(t *testing.T) {
	s := &state.State{
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha"},
		},
	}

	provision(s)

	if s.Services[0].Host != nil {
		t.Fatalf("expected nil host, got %v", s.Services[0].Host)
	}
}

func TestProvision_SkipsAlreadyProvisioned(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 5000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001"},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha", Host: &host, BindPort: &bindPort, RemotePort: &remotePort},
			{ServiceID: "svc-2", Name: "beta"},
		},
	}

	provision(s)

	if *s.Services[0].RemotePort != 5000 {
		t.Fatalf("svc-1 should still be on 5000")
	}
	if s.Services[1].RemotePort == nil || *s.Services[1].RemotePort != 5001 {
		t.Fatalf("svc-2: expected remote_port 5001, got %v", s.Services[1].RemotePort)
	}
}

func TestProvision_DeprovisionsFromShuttingDownKing(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 5000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", ShuttingDown: true},
			{BindPort: 2334, Host: "5.6.7.8", Ports: "6000-6001"},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha", Host: &host, BindPort: &bindPort, RemotePort: &remotePort, KingReady: true},
		},
	}

	provision(s)

	svc := s.Services[0]
	if svc.Host == nil || *svc.Host != "5.6.7.8" {
		t.Fatalf("expected host 5.6.7.8, got %v", svc.Host)
	}
	if svc.BindPort == nil || *svc.BindPort != 2334 {
		t.Fatalf("expected bind_port 2334, got %v", svc.BindPort)
	}
	if svc.RemotePort == nil || *svc.RemotePort != 6000 {
		t.Fatalf("expected remote_port 6000, got %v", svc.RemotePort)
	}
	if svc.KingReady {
		t.Fatalf("expected KingReady to be false after deprovisioning")
	}
}

func TestProvision_DeprovisionsAndReprovisionsToAvailableKing(t *testing.T) {
	hostA := "1.2.3.4"
	bindPortA := 2333
	remotePortA := 5000
	hostB := "5.6.7.8"
	bindPortB := 2334
	remotePortB := 6000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", ShuttingDown: true},
			{BindPort: 2334, Host: "5.6.7.8", Ports: "6000-6001"},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha", Host: &hostA, BindPort: &bindPortA, RemotePort: &remotePortA},
			{ServiceID: "svc-2", Name: "beta", Host: &hostB, BindPort: &bindPortB, RemotePort: &remotePortB},
		},
	}

	provision(s)

	svc1 := s.Services[0]
	if svc1.Host == nil || *svc1.Host != "5.6.7.8" {
		t.Fatalf("svc-1: expected host 5.6.7.8, got %v", svc1.Host)
	}
	if svc1.BindPort == nil || *svc1.BindPort != 2334 {
		t.Fatalf("svc-1: expected bind_port 2334, got %v", svc1.BindPort)
	}
	if svc1.RemotePort == nil || *svc1.RemotePort != 6001 {
		t.Fatalf("svc-1: expected remote_port 6001, got %v", svc1.RemotePort)
	}

	svc2 := s.Services[1]
	if svc2.Host == nil || *svc2.Host != "5.6.7.8" {
		t.Fatalf("svc-2: expected host 5.6.7.8 (unchanged), got %v", svc2.Host)
	}
	if svc2.RemotePort == nil || *svc2.RemotePort != 6000 {
		t.Fatalf("svc-2: expected remote_port 6000 (unchanged), got %v", svc2.RemotePort)
	}
}

func TestProvision_NoReprovisioning_WhenNoAvailableKing(t *testing.T) {
	host := "1.2.3.4"
	bindPort := 2333
	remotePort := 5000
	s := &state.State{
		Kings: []state.StateKing{
			{BindPort: 2333, Host: "1.2.3.4", Ports: "5000-5001", ShuttingDown: true},
		},
		Services: []state.StateService{
			{ServiceID: "svc-1", Name: "alpha", Host: &host, BindPort: &bindPort, RemotePort: &remotePort, KingReady: true},
		},
	}

	provision(s)

	svc := s.Services[0]
	if svc.Host != nil {
		t.Fatalf("expected nil host (unprovisioned), got %v", svc.Host)
	}
	if svc.BindPort != nil {
		t.Fatalf("expected nil bind_port, got %v", svc.BindPort)
	}
	if svc.RemotePort != nil {
		t.Fatalf("expected nil remote_port, got %v", svc.RemotePort)
	}
	if svc.KingReady {
		t.Fatalf("expected KingReady to be false")
	}
}
