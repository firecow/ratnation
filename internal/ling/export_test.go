package ling

import (
	"context"
	"net"
)

// TestUpdateTargets exposes updateTargets for external tests.
func (proxy *TCPProxy) TestUpdateTargets(targets []ProxyTarget) {
	proxy.updateTargets(targets)
}

// TestTrackUpstream exposes trackUpstream for external tests.
func (proxy *TCPProxy) TestTrackUpstream(addr string, conn net.Conn) {
	proxy.trackUpstream(addr, conn)
}

// TestUntrackUpstream exposes untrackUpstream for external tests.
func (proxy *TCPProxy) TestUntrackUpstream(addr string, conn net.Conn) {
	proxy.untrackUpstream(addr, conn)
}

// TestHandleConn exposes handleConn for external tests.
func (proxy *TCPProxy) TestHandleConn(ctx context.Context, conn net.Conn) {
	proxy.handleConn(ctx, conn)
}

// TestClose exposes close for external tests.
func (proxy *TCPProxy) TestClose() {
	proxy.close()
}

// TestSetListener sets the listener field for external tests.
func (proxy *TCPProxy) TestSetListener(listener net.Listener) {
	proxy.listener = listener
}

// TestUpstreamCount returns the number of tracked upstreams for a given address.
func (proxy *TCPProxy) TestUpstreamCount(addr string) int {
	proxy.upstreamMu.Lock()
	defer proxy.upstreamMu.Unlock()

	return len(proxy.upstreams[addr])
}

// NewProxyTarget creates a ProxyTarget for external tests.
func NewProxyTarget(host string, remotePort int) ProxyTarget {
	return ProxyTarget{host: host, remotePort: remotePort}
}
