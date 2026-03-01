package king

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/firecow/burrow/internal/state"
	"github.com/spf13/cobra"
)

type tunnelConfig struct {
	BindPort int
	Ports    string
}

func Command() *cobra.Command {
	var (
		councilHost string
		host        string
		location    string
		tunnelArgs  []string
	)

	cmd := &cobra.Command{
		Use:   "king",
		Short: "Start king",
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelConfigs, err := parseTunnelArgs(tunnelArgs)
			if err != nil {
				return err
			}
			if len(tunnelConfigs) == 0 {
				return fmt.Errorf("at least one --tunnel must be specified")
			}
			return run(cmd.Context(), councilHost, host, location, tunnelConfigs)
		},
	}

	cmd.Flags().StringVar(&councilHost, "council-host", "http://localhost:8080", "Council host to synchronize from")
	cmd.Flags().StringVar(&host, "host", "", "Host (domain or IP)")
	cmd.Flags().StringVar(&location, "location", "", "Location identifier")
	cmd.Flags().StringArrayVar(&tunnelArgs, "tunnel", nil, "Tunnel servers (bind_port=N ports=M-M)")
	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("location")

	return cmd
}

func parseTunnelArgs(args []string) ([]tunnelConfig, error) {
	configs := make([]tunnelConfig, 0, len(args))
	for _, arg := range args {
		pairs := make(map[string]string)
		for pair := range strings.SplitSeq(arg, " ") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				pairs[parts[0]] = parts[1]
			}
		}

		bindPortStr, ok := pairs["bind_port"]
		if !ok {
			return nil, fmt.Errorf("--tunnel must have 'bind_port' field")
		}
		portsStr, ok := pairs["ports"]
		if !ok {
			return nil, fmt.Errorf("--tunnel must have 'ports' field")
		}

		var bindPort int
		if _, err := fmt.Sscanf(bindPortStr, "%d", &bindPort); err != nil {
			return nil, fmt.Errorf("invalid bind_port: %s", bindPortStr)
		}

		configs = append(configs, tunnelConfig{BindPort: bindPort, Ports: portsStr})
	}
	return configs, nil
}

func run(ctx context.Context, councilHost, host, location string, tunnelConfigs []tunnelConfig) error {
	tlsCert, certPEM, err := generateSelfSignedCert()
	if err != nil {
		return fmt.Errorf("failed to generate TLS certificate: %w", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"burrow"},
		MinVersion:   tls.VersionTLS13,
	}

	tunnels := make(map[int]*tunnelServer)
	for _, r := range tunnelConfigs {
		tunnels[r.BindPort] = newTunnelServer(r.BindPort, tlsConfig)
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
		return computeReadyServiceIDs(currentState, tunnels, tunnelConfigs, host)
	}

	syncerInstance := &syncer{
		councilHost:     councilHost,
		host:            host,
		notify:          make(chan struct{}, 1),
		location:        location,
		certPEM:         certPEM,
		tunnels:         buildSyncTunnels(tunnelConfigs),
		readyServiceIDs: readyServiceIDs,
		shuttingDown: func() bool {
			mu.Lock()
			defer mu.Unlock()
			return shuttingDown
		},
	}

	for _, ts := range tunnels {
		ts.onLingConnected = syncerInstance.triggerSync
	}

	watcher := state.NewWatcher(councilHost, func(s *state.State) {
		mu.Lock()
		currentState = s
		mu.Unlock()
		onStateChanged(s, tunnels, tunnelConfigs, host)
	})

	// Start QUIC tunnel servers
	for _, ts := range tunnels {
		go func(ts *tunnelServer) {
			if err := ts.run(ctx); err != nil {
				slog.Error("Tunnel server failed", "bind_port", ts.bindPort, "error", err)
			}
		}(ts)
	}

	// Start watcher in background
	watcherCtx, watcherCancel := context.WithCancel(ctx)
	defer watcherCancel()
	go watcher.Run(watcherCtx)

	if err := watcher.WaitForState(ctx); err != nil {
		return err
	}

	// Start syncer in background
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

	// Wait for lings to notice
	time.Sleep(1 * time.Second)

	for _, ts := range tunnels {
		ts.close()
	}

	return nil
}

func buildSyncTunnels(tunnelConfigs []tunnelConfig) []syncTunnel {
	result := make([]syncTunnel, 0, len(tunnelConfigs))
	for _, r := range tunnelConfigs {
		result = append(result, syncTunnel(r))
	}
	return result
}

func computeReadyServiceIDs(s *state.State, tunnels map[int]*tunnelServer, tunnelConfigs []tunnelConfig, host string) []string {
	ids := make([]string, 0, len(s.Services))
	for _, svc := range s.Services {
		if svc.BindPort == nil || svc.Host == nil {
			continue
		}
		if *svc.Host != host {
			continue
		}

		matchesTunnel := false
		for _, r := range tunnelConfigs {
			if *svc.BindPort == r.BindPort {
				matchesTunnel = true
				break
			}
		}
		if !matchesTunnel {
			continue
		}

		lingFound := false
		for _, ling := range s.Lings {
			if ling.LingID == svc.LingID && !ling.ShuttingDown {
				lingFound = true
				break
			}
		}
		if !lingFound {
			continue
		}

		// Only report ready if the QUIC tunnel to the ling is actually established
		ts := tunnels[*svc.BindPort]
		if ts != nil {
			ts.mu.RLock()
			_, hasConn := ts.quicConns[svc.ServiceID]
			ts.mu.RUnlock()
			if !hasConn {
				continue
			}
		}

		ids = append(ids, svc.ServiceID)
	}
	return ids
}

func onStateChanged(s *state.State, tunnels map[int]*tunnelServer, tunnelConfigs []tunnelConfig, host string) {
	for _, r := range tunnelConfigs {
		ts := tunnels[r.BindPort]

		services := make(map[string]serviceAuth)
		desiredPorts := make(map[int]string) // remote_port -> service_id

		for _, svc := range s.Services {
			if svc.BindPort == nil || svc.Host == nil || svc.RemotePort == nil {
				continue
			}
			if *svc.BindPort != r.BindPort || *svc.Host != host {
				continue
			}

			lingShuttingDown := false
			for _, ling := range s.Lings {
				if ling.LingID == svc.LingID {
					lingShuttingDown = ling.ShuttingDown
					break
				}
			}
			if lingShuttingDown {
				continue
			}

			services[svc.ServiceID] = serviceAuth{token: svc.Token}
			desiredPorts[*svc.RemotePort] = svc.ServiceID
		}

		ts.updateServices(services)

		// Ensure TCP listeners for desired ports
		for port, serviceID := range desiredPorts {
			ts.ensureTCPListener(port, serviceID)
		}

		// Remove TCP listeners no longer needed
		ts.mu.RLock()
		var removePorts []int
		for port := range ts.tcpListeners {
			if _, needed := desiredPorts[port]; !needed {
				removePorts = append(removePorts, port)
			}
		}
		ts.mu.RUnlock()

		for _, port := range removePorts {
			ts.removeTCPListener(port)
		}
	}
}

func generateSelfSignedCert() (tls.Certificate, string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, "", err
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		DNSNames:              []string{"burrow"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, "", err
	}

	certPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, string(certPEMBytes), nil
}
