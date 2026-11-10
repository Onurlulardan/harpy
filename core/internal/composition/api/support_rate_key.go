package api

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

type supportRateKeyResolver struct {
	clientIP clientIPResolver
}

func newSupportRateKeyResolver(
	clientIP clientIPResolver,
) (*supportRateKeyResolver, error) {
	if clientIP == nil {
		return nil, fmt.Errorf("support client IP resolver is required")
	}
	return &supportRateKeyResolver{clientIP: clientIP}, nil
}

func (resolver *supportRateKeyResolver) Resolve(request *http.Request) ([32]byte, error) {
	if resolver == nil || resolver.clientIP == nil || request == nil {
		return [32]byte{}, fmt.Errorf("support rate evidence is unavailable")
	}
	address, err := resolver.clientIP(request)
	if err != nil || address == "" {
		return [32]byte{}, fmt.Errorf("support rate evidence is invalid")
	}
	return sha256.Sum256([]byte("harpy.support.rate.v1\x00" + address)), nil
}
