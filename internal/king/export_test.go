package king

import "github.com/quic-go/quic-go"

// SetQUICConn sets a QUIC connection for a service ID (test helper).
func (tunnelSrv *TunnelServer) SetQUICConn(serviceID string, conn *quic.Conn) {
	tunnelSrv.mu.Lock()
	defer tunnelSrv.mu.Unlock()

	tunnelSrv.quicConns[serviceID] = conn
}

// GetServiceAuth returns the service auth and whether it exists (test helper).
func (tunnelSrv *TunnelServer) GetServiceAuth(serviceID string) (ServiceAuth, bool) {
	tunnelSrv.mu.RLock()
	defer tunnelSrv.mu.RUnlock()

	auth, exists := tunnelSrv.services[serviceID]

	return auth, exists
}

// ServiceCount returns the number of registered services (test helper).
func (tunnelSrv *TunnelServer) ServiceCount() int {
	tunnelSrv.mu.RLock()
	defer tunnelSrv.mu.RUnlock()

	return len(tunnelSrv.services)
}

// SetServiceAuth sets a service auth entry (test helper).
func (tunnelSrv *TunnelServer) SetServiceAuth(serviceID string, auth ServiceAuth) {
	tunnelSrv.mu.Lock()
	defer tunnelSrv.mu.Unlock()

	tunnelSrv.services[serviceID] = auth
}
