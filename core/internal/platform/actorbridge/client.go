package actorbridge

import (
	"crypto/tls"
	"net/http"
	"sync/atomic"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type targetClient struct {
	target    Target
	transport *http.Transport
	client    *http.Client
}

type Client struct {
	targets map[actorassertionv1.Domain]targetClient
	closed  atomic.Bool
}

func NewClient(
	identity ClientIdentity,
	targets map[actorassertionv1.Domain]Target,
	now Clock,
) (*Client, error) {
	if len(identity.certificate.Certificate) < 1 || identity.certificate.PrivateKey == nil ||
		identity.roots == nil || !exactTargets(targets) || now == nil {
		return nil, NewFault(FaultInvalid)
	}
	clients := make(map[actorassertionv1.Domain]targetClient, len(targets))
	for domain, target := range targets {
		tlsConfig, err := actorassertionv1.NewClientTLSConfig(
			identity.certificate, identity.roots, target.serverName, domain, now,
		)
		if err != nil || tlsConfig.MinVersion != tls.VersionTLS13 ||
			tlsConfig.MaxVersion != tls.VersionTLS13 {
			return nil, NewFault(FaultUnavailable)
		}
		transport := &http.Transport{Proxy: nil, TLSClientConfig: tlsConfig,
			ForceAttemptHTTP2: true, DisableCompression: true,
			MaxIdleConns: len(targets), MaxIdleConnsPerHost: 1,
			IdleConnTimeout: 30 * time.Second, TLSHandshakeTimeout: 5 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second}
		client := &http.Client{Transport: transport, Timeout: 25 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return NewFault(FaultUnauthorized)
			}}
		clients[domain] = targetClient{target: target, transport: transport, client: client}
	}
	return &Client{targets: clients}, nil
}

func (client *Client) Ready() bool {
	return client != nil && !client.closed.Load() &&
		len(client.targets) == len(exactTargetDomains)
}

func (client *Client) Close() {
	if client == nil || !client.closed.CompareAndSwap(false, true) {
		return
	}
	for _, target := range client.targets {
		target.transport.CloseIdleConnections()
	}
}
