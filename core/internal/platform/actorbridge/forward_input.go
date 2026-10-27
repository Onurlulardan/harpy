package actorbridge

import (
	"crypto/sha256"
	"io"
	"net/http"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

// ForwardInput is minted from the exact replayable HTTP request. Its fields are
// opaque so a compatibility adapter cannot supply an unrelated target digest.
type ForwardInput struct {
	domain                             actorassertionv1.Domain
	resourceKey, actionKey             string
	tokenVersion, authorizationVersion int64
	targetBinding                      [32]byte
}

func NewForwardInput(
	domain actorassertionv1.Domain,
	resourceKey string,
	actionKey string,
	tokenVersion int64,
	authorizationVersion int64,
	request *http.Request,
) (ForwardInput, error) {
	binding, err := forwardRequestBinding(request)
	input := ForwardInput{domain: domain, resourceKey: resourceKey,
		actionKey: actionKey, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, targetBinding: binding}
	if err != nil || input.issueInput().valid() == false {
		return ForwardInput{}, NewFault(FaultInvalid)
	}
	return input, nil
}

func (input ForwardInput) issueInput() IssueInput {
	return IssueInput{TargetDomain: input.domain, ResourceKey: input.resourceKey,
		ActionKey: input.actionKey, TargetBinding: input.targetBinding,
		TokenVersion:         input.tokenVersion,
		AuthorizationVersion: input.authorizationVersion}
}

func (input ForwardInput) validateRequest(request *http.Request) (IssueInput, error) {
	binding, err := forwardRequestBinding(request)
	issue := input.issueInput()
	if err != nil || !issue.valid() || binding != input.targetBinding {
		return IssueInput{}, NewFault(FaultInvalid)
	}
	return issue, nil
}

func forwardRequestBinding(request *http.Request) ([32]byte, error) {
	if !validForwardRequestTarget(request) {
		return [32]byte{}, NewFault(FaultInvalid)
	}
	bodyDigest, err := forwardRequestBodyDigest(request)
	if err != nil {
		return [32]byte{}, err
	}
	binding, err := actorassertionv1.HTTPRequestTargetBindingSHA256(
		request.Method, request.URL.EscapedPath(), request.URL.RawQuery,
		bodyDigest, request.ContentLength,
	)
	if err != nil {
		return [32]byte{}, NewFault(FaultInvalid)
	}
	return binding, nil
}

func validForwardRequestTarget(request *http.Request) bool {
	if request == nil || request.URL == nil || request.URL.Scheme != "" ||
		request.URL.Host != "" || request.URL.User != nil || request.URL.Opaque != "" ||
		request.URL.Path == "" || request.URL.Path[0] != '/' || request.URL.RawPath != "" ||
		request.URL.Fragment != "" ||
		len(request.URL.EscapedPath()) > actorassertionv1.MaximumHTTPBindingPathBytes ||
		len(request.URL.RawQuery) > actorassertionv1.MaximumHTTPBindingQueryBytes {
		return false
	}
	switch request.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func forwardRequestBodyDigest(request *http.Request) ([32]byte, error) {
	if request.ContentLength < 0 ||
		request.ContentLength > actorassertionv1.MaximumHTTPBindingBodyBytes {
		return [32]byte{}, NewFault(FaultInvalid)
	}
	if request.Body == nil || request.Body == http.NoBody {
		if request.ContentLength != 0 {
			return [32]byte{}, NewFault(FaultInvalid)
		}
		return sha256.Sum256(nil), nil
	}
	if request.GetBody == nil {
		return [32]byte{}, NewFault(FaultInvalid)
	}
	body, err := request.GetBody()
	if err != nil || body == nil {
		return [32]byte{}, NewFault(FaultInvalid)
	}
	defer body.Close()
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(
		body, actorassertionv1.MaximumHTTPBindingBodyBytes+1,
	))
	if err != nil || written != request.ContentLength {
		return [32]byte{}, NewFault(FaultInvalid)
	}
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}
