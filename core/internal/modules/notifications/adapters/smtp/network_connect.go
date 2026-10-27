package smtp

import (
	"context"
	"fmt"
	"net"
	"strconv"
)

func (exchange *networkExchanger) dial(ctx context.Context,
	host string, port int,
) (net.Conn, error) {
	addresses, err := exchange.resolver.LookupIPAddr(ctx, host)
	if err != nil || len(addresses) == 0 {
		return nil, fmt.Errorf("resolve SMTP endpoint")
	}
	for _, address := range addresses {
		if !safeProviderIP(address.IP) {
			return nil, fmt.Errorf("unsafe SMTP endpoint")
		}
	}
	target := net.JoinHostPort(addresses[0].IP.String(), strconv.Itoa(port))
	return exchange.dialer.DialContext(ctx, "tcp", target)
}

func safeProviderIP(address net.IP) bool {
	return address != nil && !address.IsPrivate() && !address.IsLoopback() &&
		!address.IsLinkLocalUnicast() && !address.IsLinkLocalMulticast() &&
		!address.IsUnspecified() && !address.IsMulticast()
}
