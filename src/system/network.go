package system

import (
	"context"
	"net"
	"time"
)

const networkProbeTimeout = 2 * time.Second

// HasInternetAccess checks whether external packet routing is available without blocking indefinitely.
func HasInternetAccess(ctx context.Context) bool {
	targets := []string{"1.1.1.1:53", "8.8.8.8:53"}
	dialer := &net.Dialer{Timeout: networkProbeTimeout}

	for _, target := range targets {
		conn, err := dialer.DialContext(ctx, "tcp", target)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}
