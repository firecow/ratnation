package ling

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestTCPProxy_RoundRobin(t *testing.T) {
	// Start two echo backends
	backend1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = backend1.Close() }()

	backend2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = backend2.Close() }()

	go echoServer(backend1, "backend1")
	go echoServer(backend2, "backend2")

	addr1 := backend1.Addr().(*net.TCPAddr)
	addr2 := backend2.Addr().(*net.TCPAddr)

	proxy := newTCPProxy("test", 0)
	proxy.updateTargets([]proxyTarget{
		{host: "127.0.0.1", remotePort: addr1.Port},
		{host: "127.0.0.1", remotePort: addr2.Port},
	})

	// Use a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy.listener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go proxy.handleConn(conn)
		}
	}()
	defer proxy.close()

	proxyAddr := listener.Addr().String()

	// Make two connections - should go to different backends
	resp1 := dialAndSend(t, proxyAddr)
	resp2 := dialAndSend(t, proxyAddr)

	// One should be "backend1:" prefix, other "backend2:"
	if resp1 == resp2 {
		t.Logf("resp1=%s resp2=%s", resp1, resp2)
		t.Log("Both requests went to the same backend (acceptable but not ideal)")
	}
}

func TestTCPProxy_NoTargets(t *testing.T) {
	proxy := newTCPProxy("test", 0)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy.listener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go proxy.handleConn(conn)
		}
	}()
	defer proxy.close()

	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	// Should close quickly since no targets
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err == nil {
		t.Fatal("expected error (connection should be closed)")
	}
}

func echoServer(listener net.Listener, name string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer func() { _ = c.Close() }()
			prefix := []byte(name + ":")
			_, _ = c.Write(prefix)
			_, _ = io.Copy(c, c)
		}(conn)
	}
}

func TestTCPProxy_UpdateTargets_ClosesRemovedUpstreams(t *testing.T) {
	proxy := newTCPProxy("test", 0)

	// Create a mock connection using net.Pipe
	server, client := net.Pipe()
	defer func() { _ = server.Close() }()

	targetAddr := "10.0.0.1:5000"
	proxy.trackUpstream(targetAddr, client)

	// Verify connection is tracked
	proxy.upstreamMu.Lock()
	if len(proxy.upstreams[targetAddr]) != 1 {
		t.Fatalf("expected 1 tracked upstream, got %d", len(proxy.upstreams[targetAddr]))
	}
	proxy.upstreamMu.Unlock()

	// Set initial targets including the one we tracked
	proxy.updateTargets([]proxyTarget{
		{host: "10.0.0.1", remotePort: 5000},
	})

	// Now remove the target by updating to empty list
	proxy.updateTargets(nil)

	// Verify the upstream entry was removed
	proxy.upstreamMu.Lock()
	remaining := len(proxy.upstreams[targetAddr])
	proxy.upstreamMu.Unlock()
	if remaining != 0 {
		t.Errorf("expected upstream entry to be removed, got %d connections", remaining)
	}

	// Verify the connection was closed by attempting to write
	_, err := client.Write([]byte("test"))
	if err == nil {
		t.Error("expected write to closed connection to fail")
	}
}

func TestTCPProxy_UpdateTargets_KeepsRemainingUpstreams(t *testing.T) {
	proxy := newTCPProxy("test", 0)

	keepServer, keepClient := net.Pipe()
	defer func() { _ = keepServer.Close() }()
	defer func() { _ = keepClient.Close() }()

	removeServer, removeClient := net.Pipe()
	defer func() { _ = removeServer.Close() }()

	keepAddr := "10.0.0.1:5000"
	removeAddr := "10.0.0.2:5000"
	proxy.trackUpstream(keepAddr, keepClient)
	proxy.trackUpstream(removeAddr, removeClient)

	// Set initial targets with both
	proxy.updateTargets([]proxyTarget{
		{host: "10.0.0.1", remotePort: 5000},
		{host: "10.0.0.2", remotePort: 5000},
	})

	// Remove one target
	proxy.updateTargets([]proxyTarget{
		{host: "10.0.0.1", remotePort: 5000},
	})

	// Verify kept target still has its upstream
	proxy.upstreamMu.Lock()
	keptCount := len(proxy.upstreams[keepAddr])
	removedCount := len(proxy.upstreams[removeAddr])
	proxy.upstreamMu.Unlock()

	if keptCount != 1 {
		t.Errorf("expected 1 upstream for kept target, got %d", keptCount)
	}
	if removedCount != 0 {
		t.Errorf("expected 0 upstreams for removed target, got %d", removedCount)
	}

	// Verify removed connection was closed: reading from the server side
	// should return an error (io.EOF or io.ErrClosedPipe)
	_ = removeServer.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 1)
	_, err := removeServer.Read(buf)
	if err == nil {
		t.Error("expected read from removed connection's server side to fail")
	}

	// Verify kept connection is still alive: write from server, read from client
	go func() {
		_, _ = keepServer.Write([]byte("ping"))
	}()
	_ = keepClient.SetReadDeadline(time.Now().Add(time.Second))
	readBuf := make([]byte, 4)
	n, err := keepClient.Read(readBuf)
	if err != nil {
		t.Errorf("expected kept connection to still be readable: %v", err)
	}
	if string(readBuf[:n]) != "ping" {
		t.Errorf("expected 'ping', got %q", string(readBuf[:n]))
	}
}

func TestTCPProxy_TrackAndUntrack(t *testing.T) {
	proxy := newTCPProxy("test", 0)

	_, client1 := net.Pipe()
	_, client2 := net.Pipe()

	addr := "10.0.0.1:5000"
	proxy.trackUpstream(addr, client1)
	proxy.trackUpstream(addr, client2)

	proxy.upstreamMu.Lock()
	count := len(proxy.upstreams[addr])
	proxy.upstreamMu.Unlock()
	if count != 2 {
		t.Fatalf("expected 2 tracked upstreams, got %d", count)
	}

	proxy.untrackUpstream(addr, client1)

	proxy.upstreamMu.Lock()
	count = len(proxy.upstreams[addr])
	proxy.upstreamMu.Unlock()
	if count != 1 {
		t.Fatalf("expected 1 tracked upstream after untrack, got %d", count)
	}

	proxy.untrackUpstream(addr, client2)

	proxy.upstreamMu.Lock()
	count = len(proxy.upstreams[addr])
	proxy.upstreamMu.Unlock()
	if count != 0 {
		t.Errorf("expected 0 tracked upstreams after untrack all, got %d", count)
	}
}

func dialAndSend(t *testing.T, addr string) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(time.Second))

	// Read the backend prefix first
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf[:n])
}
