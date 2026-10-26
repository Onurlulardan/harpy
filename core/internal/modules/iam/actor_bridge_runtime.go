package iam

import (
	"context"
	"net/http"

	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ActorBridgeRuntime struct {
	forwarder *platformactorbridge.Forwarder
}

func (runtime *ActorBridgeRuntime) Ready() bool {
	return runtime != nil && runtime.forwarder != nil && runtime.forwarder.Ready()
}

func (runtime *ActorBridgeRuntime) Close() {
	if runtime == nil {
		return
	}
	if runtime.forwarder != nil {
		runtime.forwarder.Close()
	}
}

func (runtime *ActorBridgeRuntime) IssueAndForward(
	ctx context.Context,
	scopeRequest platformtx.ScopeRequest,
	input platformactorbridge.ForwardInput,
	request *http.Request,
) (*http.Response, error) {
	if !runtime.Ready() {
		return nil, platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	return runtime.forwarder.IssueAndForward(ctx, scopeRequest, input, request)
}
