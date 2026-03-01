package ling

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/firecow/burrow/internal/state"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type tunnelConfig struct {
	Name      string
	LocalAddr string
}

type proxyConfig struct {
	Name     string
	BindPort int
}

func Command() *cobra.Command {
	var (
		councilHost       string
		lingID            string
		tunnelArgs        []string
		proxyArgs         []string
		preferredLocation string
	)

	cmd := &cobra.Command{
		Use:   "ling",
		Short: "Start ling",
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelConfigs, err := parseTunnelArgs(tunnelArgs)
			if err != nil {
				return err
			}
			proxies, err := parseProxyArgs(proxyArgs)
			if err != nil {
				return err
			}
			if lingID == "" {
				lingID = uuid.New().String()
			}
			return run(cmd.Context(), councilHost, lingID, preferredLocation, tunnelConfigs, proxies)
		},
	}

	cmd.Flags().StringVar(&councilHost, "council-host", "http://localhost:8080", "Council host to synchronize from")
	cmd.Flags().StringVar(&lingID, "ling-id", "", "Unique ID of this ling instance (auto-generates UUID if omitted)")
	cmd.Flags().StringArrayVar(&tunnelArgs, "tunnel", nil, "Tunnel clients (name=STR local_addr=ADDR)")
	cmd.Flags().StringArrayVar(&proxyArgs, "proxy", nil, "TCP proxies (name=STR bind_port=N)")
	cmd.Flags().StringVar(&preferredLocation, "location", "default", "Preferred location identifier")

	return cmd
}

func parseTunnelArgs(args []string) ([]tunnelConfig, error) {
	if len(args) == 0 {
		return nil, nil
	}
	configs := make([]tunnelConfig, 0, len(args))
	for _, arg := range args {
		pairs := make(map[string]string)
		for pair := range strings.SplitSeq(arg, " ") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				pairs[parts[0]] = parts[1]
			}
		}

		name, ok := pairs["name"]
		if !ok {
			return nil, fmt.Errorf("--tunnel must have 'name' field")
		}
		localAddr, ok := pairs["local_addr"]
		if !ok {
			return nil, fmt.Errorf("--tunnel must have 'local_addr' field")
		}

		configs = append(configs, tunnelConfig{Name: name, LocalAddr: localAddr})
	}
	return configs, nil
}

func parseProxyArgs(args []string) ([]proxyConfig, error) {
	if len(args) == 0 {
		return nil, nil
	}
	configs := make([]proxyConfig, 0, len(args))
	for _, arg := range args {
		pairs := make(map[string]string)
		for pair := range strings.SplitSeq(arg, " ") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				pairs[parts[0]] = parts[1]
			}
		}

		name, ok := pairs["name"]
		if !ok {
			return nil, fmt.Errorf("--proxy must have 'name' field")
		}
		bindPortStr, ok := pairs["bind_port"]
		if !ok {
			return nil, fmt.Errorf("--proxy must have 'bind_port' field")
		}
		bindPort, err := strconv.Atoi(bindPortStr)
		if err != nil {
			return nil, fmt.Errorf("invalid bind_port: %s", bindPortStr)
		}

		configs = append(configs, proxyConfig{Name: name, BindPort: bindPort})
	}
	return configs, nil
}

func run(ctx context.Context, councilHost, lingID, preferredLocation string, tunnelConfigs []tunnelConfig, proxies []proxyConfig) error {
	// Build tunnel name -> local_addr map
	tunnelMap := make(map[string]string)
	for _, r := range tunnelConfigs {
		tunnelMap[r.Name] = r.LocalAddr
	}

	// Only tunnel services get registered with council (not proxy-only names)
	syncTunnels := make([]syncTunnel, 0, len(tunnelConfigs))
	for _, r := range tunnelConfigs {
		syncTunnels = append(syncTunnels, syncTunnel{Name: r.Name})
	}
	if len(syncTunnels) == 0 {
		syncTunnels = []syncTunnel{}
	}

	tunnelClient := newTunnelClient()
	tcpProxies := make(map[string]*tcpProxy)

	for _, p := range proxies {
		proxy := newTCPProxy(p.Name, p.BindPort)
		tcpProxies[p.Name] = proxy
		if err := proxy.start(); err != nil {
			return fmt.Errorf("failed to start proxy %s: %w", p.Name, err)
		}
	}

	var mu sync.Mutex
	var currentState *state.State
	shuttingDown := false

	readyServiceIDs := func() []string {
		mu.Lock()
		defer mu.Unlock()
		if currentState == nil {
			return []string{}
		}
		return computeReadyServiceIDs(currentState, lingID, tunnelMap)
	}

	syncerInstance := &lingSyncer{
		councilHost:       councilHost,
		lingID:            lingID,
		preferredLocation: preferredLocation,
		tunnels:           syncTunnels,
		readyServiceIDs:   readyServiceIDs,
		notify:            make(chan struct{}, 1),
		shuttingDown: func() bool {
			mu.Lock()
			defer mu.Unlock()
			return shuttingDown
		},
	}

	tunnelClient.onConnected = syncerInstance.triggerSync

	watcher := state.NewWatcher(councilHost, func(s *state.State) {
		mu.Lock()
		currentState = s
		mu.Unlock()
		onStateChanged(ctx, s, lingID, tunnelMap, tunnelClient, tcpProxies)
	})

	watcherCtx, watcherCancel := context.WithCancel(ctx)
	defer watcherCancel()
	go watcher.Run(watcherCtx)

	if err := watcher.WaitForState(ctx); err != nil {
		return err
	}

	syncerCtx, syncerCancel := context.WithCancel(ctx)
	defer syncerCancel()
	go syncerInstance.run(syncerCtx)

	slog.Info("Ready")

	<-ctx.Done()

	slog.Info("Shutdown sequence initiated")
	mu.Lock()
	shuttingDown = true
	mu.Unlock()

	watcherCancel()
	syncerCancel()

	// Final sync to broadcast shutting_down=true
	finalCtx, finalCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer finalCancel()
	syncerInstance.sync(finalCtx)

	// Wait for kings to notice
	time.Sleep(750 * time.Millisecond)

	tunnelClient.closeAll()
	for _, proxy := range tcpProxies {
		proxy.close()
	}

	return nil
}

func computeReadyServiceIDs(s *state.State, lingID string, tunnelMap map[string]string) []string {
	ids := make([]string, 0, len(s.Services))
	for _, svc := range s.Services {
		if svc.LingID != lingID {
			continue
		}
		if _, ok := tunnelMap[svc.Name]; !ok {
			continue
		}
		if svc.KingReady {
			ids = append(ids, svc.ServiceID)
		}
	}
	return ids
}

func onStateChanged(ctx context.Context, s *state.State, lingID string, tunnelMap map[string]string, tc *tunnelClient, tcpProxies map[string]*tcpProxy) {
	kings := buildKingIndex(s)
	updateTunnelConnections(ctx, s, lingID, tunnelMap, tc, kings)
	updateProxyTargets(s, tcpProxies, kings)
}

type kingGroup struct {
	host     string
	bindPort int
	certPEM  string
	services []tunnelService
}

type kingIndex struct {
	healthy bool
	certPEM string
}

func buildKingIndex(s *state.State) map[string]kingIndex {
	index := make(map[string]kingIndex, len(s.Kings))
	for _, king := range s.Kings {
		addr := net.JoinHostPort(king.Host, strconv.Itoa(king.BindPort))
		index[addr] = kingIndex{healthy: !king.ShuttingDown, certPEM: king.CertPEM}
	}
	return index
}

func buildHealthyLingSet(s *state.State) map[string]bool {
	set := make(map[string]bool, len(s.Lings))
	for _, ling := range s.Lings {
		if !ling.ShuttingDown {
			set[ling.LingID] = true
		}
	}
	return set
}

func updateTunnelConnections(ctx context.Context, s *state.State, lingID string, tunnelMap map[string]string, tc *tunnelClient, kings map[string]kingIndex) {
	groups := make(map[string]*kingGroup)
	localAddrs := make(map[string]string)

	for _, svc := range s.Services {
		if svc.LingID != lingID {
			continue
		}
		localAddr, isTunnel := tunnelMap[svc.Name]
		if !isTunnel || svc.BindPort == nil || svc.Host == nil {
			continue
		}

		kingAddr := net.JoinHostPort(*svc.Host, strconv.Itoa(*svc.BindPort))
		king, exists := kings[kingAddr]
		if !exists || !king.healthy {
			continue
		}

		if groups[kingAddr] == nil {
			groups[kingAddr] = &kingGroup{
				host:     *svc.Host,
				bindPort: *svc.BindPort,
				certPEM:  king.certPEM,
			}
		}
		groups[kingAddr].services = append(groups[kingAddr].services, tunnelService{
			serviceID: svc.ServiceID,
			token:     svc.Token,
			localAddr: localAddr,
		})
		localAddrs[svc.ServiceID] = localAddr
	}

	for _, group := range groups {
		tc.ensureConnection(ctx, group, localAddrs)
	}

	tc.mu.Lock()
	for addr := range tc.connections {
		if _, needed := groups[addr]; !needed {
			_ = tc.connections[addr].CloseWithError(0, "no longer needed")
			delete(tc.connections, addr)
		}
	}
	tc.mu.Unlock()
}

func updateProxyTargets(s *state.State, tcpProxies map[string]*tcpProxy, kings map[string]kingIndex) {
	healthyLings := buildHealthyLingSet(s)

	for proxyName, proxy := range tcpProxies {
		targets := make([]proxyTarget, 0, len(s.Services))
		for _, svc := range s.Services {
			if svc.Name != proxyName || svc.Host == nil || svc.RemotePort == nil {
				continue
			}
			if !svc.LingReady || !svc.KingReady {
				continue
			}
			if !healthyLings[svc.LingID] {
				continue
			}
			if svc.BindPort == nil {
				continue
			}
			kingAddr := net.JoinHostPort(*svc.Host, strconv.Itoa(*svc.BindPort))
			if king, exists := kings[kingAddr]; !exists || !king.healthy {
				continue
			}

			targets = append(targets, proxyTarget{
				host:       *svc.Host,
				remotePort: *svc.RemotePort,
			})
		}
		proxy.updateTargets(targets)
	}
}
