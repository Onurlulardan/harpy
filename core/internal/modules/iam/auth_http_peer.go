package iam

import (
	"context"
	"net"
	"net/netip"
)

type authHTTPPeerContextKey struct{}

type authHTTPPeer struct {
	ip string
}

func withAuthHTTPPeer(ctx context.Context, remoteAddress string) context.Context {
	if ctx == nil {
		return nil
	}
	ip, valid := canonicalAuthHTTPPeerIP(remoteAddress)
	if !valid {
		return ctx
	}
	return context.WithValue(ctx, authHTTPPeerContextKey{}, authHTTPPeer{ip: ip})
}

func authHTTPPeerIP(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	peer, valid := ctx.Value(authHTTPPeerContextKey{}).(authHTTPPeer)
	return peer.ip, valid && peer.ip != ""
}

func canonicalAuthHTTPPeerIP(remoteAddress string) (string, bool) {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = remoteAddress
	}
	address, err := netip.ParseAddr(host)
	if err != nil || !address.IsValid() || address.IsUnspecified() {
		return "", false
	}
	address = address.WithZone("").Unmap()
	return address.String(), true
}
