package ling

import (
	"io"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type proxyTarget struct {
	host       string
	remotePort int
}

func proxyTargetAddr(t proxyTarget) string {
	return net.JoinHostPort(t.host, strconv.Itoa(t.remotePort))
}

type tcpProxy struct {
	name       string
	bindPort   int
	mu         sync.RWMutex
	targets    []proxyTarget
	counter    atomic.Uint64
	listener   net.Listener
	closeOnce  sync.Once
	hasTargets chan struct{} // closed when targets are available, reset when empty

	upstreamMu sync.Mutex
	upstreams  map[string][]net.Conn // target addr -> active upstream connections
}

func newTCPProxy(name string, bindPort int) *tcpProxy {
	return &tcpProxy{
		name:       name,
		bindPort:   bindPort,
		upstreams:  make(map[string][]net.Conn),
		hasTargets: make(chan struct{}),
	}
}

func (p *tcpProxy) updateTargets(targets []proxyTarget) {
	p.mu.Lock()
	oldTargets := p.targets
	p.targets = targets
	if len(targets) > 0 {
		select {
		case <-p.hasTargets:
		default:
			close(p.hasTargets)
		}
	} else {
		select {
		case <-p.hasTargets:
			p.hasTargets = make(chan struct{})
		default:
		}
	}
	p.mu.Unlock()

	// Find removed targets and close their upstream connections
	newAddrs := make(map[string]bool)
	for _, t := range targets {
		newAddrs[proxyTargetAddr(t)] = true
	}

	p.upstreamMu.Lock()
	for _, t := range oldTargets {
		addr := proxyTargetAddr(t)
		if !newAddrs[addr] {
			for _, conn := range p.upstreams[addr] {
				conn.Close()
			}
			delete(p.upstreams, addr)
		}
	}
	p.upstreamMu.Unlock()
}

func (p *tcpProxy) trackUpstream(addr string, conn net.Conn) {
	p.upstreamMu.Lock()
	p.upstreams[addr] = append(p.upstreams[addr], conn)
	p.upstreamMu.Unlock()
}

func (p *tcpProxy) untrackUpstream(addr string, conn net.Conn) {
	p.upstreamMu.Lock()
	conns := p.upstreams[addr]
	for i, c := range conns {
		if c == conn {
			p.upstreams[addr] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	p.upstreamMu.Unlock()
}

func (p *tcpProxy) start() error {
	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(p.bindPort)))
	if err != nil {
		return err
	}
	p.listener = listener

	slog.Info("TCP proxy started", "name", p.name, "bind_port", p.bindPort)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go p.handleConn(conn)
		}
	}()

	return nil
}

func (p *tcpProxy) handleConn(clientConn net.Conn) {
	defer clientConn.Close()

	deadline := time.After(30 * time.Second)

	var targets []proxyTarget
	for {
		p.mu.RLock()
		targets = p.targets
		hasTargets := p.hasTargets
		p.mu.RUnlock()

		if len(targets) > 0 {
			break
		}

		select {
		case <-hasTargets:
			continue
		case <-deadline:
			return
		}
	}

	// Round-robin
	idx := p.counter.Add(1) - 1
	target := targets[idx%uint64(len(targets))]

	addr := proxyTargetAddr(target)
	upstream, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Error("Failed to dial upstream", "name", p.name, "addr", addr, "error", err)
		return
	}
	defer upstream.Close()

	p.trackUpstream(addr, upstream)
	defer p.untrackUpstream(addr, upstream)

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(upstream, clientConn)
		close(done)
	}()
	_, _ = io.Copy(clientConn, upstream)
	<-done
}

func (p *tcpProxy) close() {
	p.closeOnce.Do(func() {
		if p.listener != nil {
			p.listener.Close()
		}
		p.mu.Lock()
		select {
		case <-p.hasTargets:
		default:
			close(p.hasTargets)
		}
		p.mu.Unlock()
	})
}
