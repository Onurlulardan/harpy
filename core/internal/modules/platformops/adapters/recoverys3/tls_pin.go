package recoverys3

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"net"
	"net/http"
	"time"
)

func pinnedHTTPClient(scheme, digest string) (*http.Client, error) {
	expected, err := hex.DecodeString(digest)
	if err != nil || len(expected) != sha256.Size {
		return nil, ErrObject
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyFromEnvironment
	transport.DialContext = (&net.Dialer{
		Timeout: 10 * time.Second, KeepAlive: 30 * time.Second,
	}).DialContext
	transport.ResponseHeaderTimeout = 30 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second
	if scheme == "https" {
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			VerifyConnection: func(state tls.ConnectionState) error {
				for _, chains := range state.VerifiedChains {
					for _, certificate := range chains {
						actual := sha256.Sum256(certificate.Raw)
						if string(actual[:]) == string(expected) {
							return nil
						}
					}
				}
				return ErrObject
			},
		}
	}
	return &http.Client{Transport: transport, Timeout: 0}, nil
}
