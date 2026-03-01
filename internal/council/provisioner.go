package council

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/firecow/ratnation/internal/state"
)

type availableKingPort struct {
	king  *state.StateKing
	ports []int
}

func availableKingPorts(s *state.State) []availableKingPort {
	var result []availableKingPort

	for i := range s.Kings {
		king := &s.Kings[i]
		if king.ShuttingDown {
			continue
		}

		parts := strings.SplitN(king.Ports, "-", 2)
		if len(parts) != 2 {
			continue
		}
		from, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		to, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		used := make(map[int]bool)
		for _, svc := range s.Services {
			if svc.BindPort != nil && svc.Host != nil && *svc.BindPort == king.BindPort && *svc.Host == king.Host {
				if svc.RemotePort != nil {
					used[*svc.RemotePort] = true
				}
			}
		}

		var ports []int
		for port := from; port <= to; port++ {
			if !used[port] {
				ports = append(ports, port)
			}
		}

		if len(ports) > 0 {
			result = append(result, availableKingPort{king: king, ports: ports})
		}
	}

	return result
}

func provisionService(s *state.State, service *state.StateService) {
	available := availableKingPorts(s)
	if len(available) == 0 {
		slog.Error("No available remote_port found on any kings", "service_id", service.ServiceID)
		return
	}

	first := available[0]
	remotePort := first.ports[0]

	service.Host = &first.king.Host
	service.RemotePort = &remotePort
	service.BindPort = &first.king.BindPort

	s.Revision++

	slog.Info("Provisioned service",
		"name", service.Name,
		"host", first.king.Host,
		"bind_port", first.king.BindPort,
		"remote_port", remotePort,
	)
}

func provision(s *state.State) {
	// Deprovision services from shutting-down kings
	for i := range s.Services {
		svc := &s.Services[i]
		if svc.BindPort == nil || svc.Host == nil {
			continue
		}
		for _, king := range s.Kings {
			if king.Host == *svc.Host && king.BindPort == *svc.BindPort && king.ShuttingDown {
				slog.Info("Deprovisioning service from shutting-down king", "name", svc.Name, "service_id", svc.ServiceID)
				svc.Host = nil
				svc.BindPort = nil
				svc.RemotePort = nil
				svc.KingReady = false
				break
			}
		}
	}

	for i := range s.Services {
		if s.Services[i].BindPort == nil {
			provisionService(s, &s.Services[i])
		}
	}
}
