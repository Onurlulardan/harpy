package actorbridge

import "sync/atomic"

// Forwarder is the internal mixed-domain boundary. It owns issuance and the
// outbound mTLS client so callers cannot send an unsigned legacy request.
type Forwarder struct {
	issuer *Issuer
	client *Client
	closed atomic.Bool
}

func NewForwarder(issuer *Issuer, client *Client) (*Forwarder, error) {
	if issuer == nil || !issuer.Ready() || client == nil || !client.Ready() {
		return nil, NewFault(FaultInvalid)
	}
	return &Forwarder{issuer: issuer, client: client}, nil
}

func (forwarder *Forwarder) Ready() bool {
	return forwarder != nil && !forwarder.closed.Load() &&
		forwarder.issuer != nil && forwarder.issuer.Ready() &&
		forwarder.client != nil && forwarder.client.Ready()
}

func (forwarder *Forwarder) Close() {
	if forwarder == nil || !forwarder.closed.CompareAndSwap(false, true) {
		return
	}
	forwarder.client.Close()
	forwarder.issuer.Close()
}
