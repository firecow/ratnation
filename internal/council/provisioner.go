package council

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/firecow/burrow/internal/state"
)

const portRangeParts = 2

// AvailableKingPort holds a king and its available ports.
type AvailableKingPort struct {
	King  *state.King
	Ports []int
}

// AvailableKingPorts returns all kings with their available (unused) ports.
func AvailableKingPorts(currentState *state.State) []AvailableKingPort {
	var result []AvailableKingPort

	for i := range currentState.Kings {
		king := &currentState.Kings[i]

		if king.ShuttingDown {
			continue
		}

		ports := parseKingPorts(king, currentState)
		if len(ports) > 0 {
			result = append(result, AvailableKingPort{King: king, Ports: ports})
		}
	}

	return result
}

func parseKingPorts(king *state.King, currentState *state.State) []int {
	parts := strings.SplitN(king.Ports, "-", portRangeParts)
	if len(parts) != portRangeParts {
		return nil
	}

	from, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}

	portEnd, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil
	}

	used := collectUsedPorts(king, currentState)

	var ports []int

	for port := from; port <= portEnd; port++ {
		if !used[port] {
			ports = append(ports, port)
		}
	}

	return ports
}

func collectUsedPorts(king *state.King, currentState *state.State) map[int]bool {
	used := make(map[int]bool)

	for _, svc := range currentState.Services {
		if svc.BindPort != nil &&
			svc.Host != nil &&
			*svc.BindPort == king.BindPort &&
			*svc.Host == king.Host {
			if svc.RemotePort != nil {
				used[*svc.RemotePort] = true
			}
		}
	}

	return used
}

// ProvisionService assigns the first available king port to a service.
func ProvisionService(currentState *state.State, service *state.Service) {
	available := AvailableKingPorts(currentState)
	if len(available) == 0 {
		slog.Error("No available remote_port found on any kings", "service_id", service.ServiceID)

		return
	}

	first := available[0]
	remotePort := first.Ports[0]

	service.Host = &first.King.Host
	service.RemotePort = &remotePort
	service.BindPort = &first.King.BindPort

	currentState.Revision++

	slog.Info("Provisioned service",
		"name", service.Name,
		"host", first.King.Host,
		"bind_port", first.King.BindPort,
		"remote_port", remotePort,
	)
}

// Provision deprovisions services from shutting-down kings and provisions
// unprovisioned services.
func Provision(currentState *state.State) {
	deprovisionFromShuttingDownKings(currentState)

	for i := range currentState.Services {
		if currentState.Services[i].BindPort == nil {
			ProvisionService(currentState, &currentState.Services[i])
		}
	}
}

func deprovisionFromShuttingDownKings(currentState *state.State) {
	for i := range currentState.Services {
		svc := &currentState.Services[i]

		if svc.BindPort == nil || svc.Host == nil {
			continue
		}

		for _, king := range currentState.Kings {
			if king.Host == *svc.Host && king.BindPort == *svc.BindPort && king.ShuttingDown {
				slog.Info(
					"Deprovisioning service from shutting-down king",
					"name", svc.Name,
					"service_id", svc.ServiceID,
				)

				svc.Host = nil
				svc.BindPort = nil
				svc.RemotePort = nil
				svc.KingReady = false

				break
			}
		}
	}
}
