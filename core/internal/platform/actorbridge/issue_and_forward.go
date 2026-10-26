package actorbridge

import (
	"context"
	"net/http"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// IssueAndForward commits current IAM and cutover evidence before opening the
// network hop. The target remains responsible for consuming the assertion and
// applying its business effect in one caller-owned V1 transaction.
func (forwarder *Forwarder) IssueAndForward(
	ctx context.Context,
	scopeRequest platformtx.ScopeRequest,
	input ForwardInput,
	request *http.Request,
) (*http.Response, error) {
	if !forwarder.Ready() || ctx == nil || ctx.Err() != nil || request == nil ||
		!platformtx.NetworkAllowed(ctx) {
		return nil, NewFault(FaultInvalid)
	}
	issue, err := input.validateRequest(request)
	if err != nil {
		return nil, err
	}
	credential, err := forwarder.issuer.Issue(ctx, scopeRequest, issue)
	if err != nil {
		return nil, err
	}
	if !forwarder.Ready() {
		return nil, NewFault(FaultUnavailable)
	}
	prepared, err := replayForwardRequest(request)
	if err != nil {
		return nil, err
	}
	return forwarder.client.Do(ctx, issue.TargetDomain, credential, prepared)
}

func replayForwardRequest(request *http.Request) (*http.Request, error) {
	prepared := request.Clone(request.Context())
	if request.Body == nil || request.Body == http.NoBody {
		prepared.Body = http.NoBody
		return prepared, nil
	}
	body, err := request.GetBody()
	if err != nil || body == nil {
		return nil, NewFault(FaultInvalid)
	}
	prepared.Body = body
	return prepared, nil
}
