package actorbridge

import (
	"context"
	"net/http"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (client *Client) Do(
	ctx context.Context,
	domain actorassertionv1.Domain,
	credential actorassertionv1.Credential,
	request *http.Request,
) (*http.Response, error) {
	if ctx == nil || ctx.Err() != nil || !platformtx.NetworkAllowed(ctx) {
		return nil, NewFault(FaultInvalid)
	}
	prepared, err := client.Prepare(request, domain, credential)
	if err != nil {
		return nil, err
	}
	target, exists := client.targets[domain]
	if !exists || client.closed.Load() {
		return nil, NewFault(FaultUnavailable)
	}
	response, err := target.client.Do(prepared.WithContext(ctx))
	if err != nil {
		return nil, runtimeFault(err, FaultUnavailable)
	}
	return response, nil
}
