package actorbridge

import (
	"net/http"
	"net/url"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func (client *Client) Prepare(
	request *http.Request,
	domain actorassertionv1.Domain,
	credential actorassertionv1.Credential,
) (*http.Request, error) {
	if !client.Ready() || request == nil || request.URL == nil ||
		request.Method == "" || request.URL.Scheme != "" || request.URL.Host != "" ||
		request.URL.Opaque != "" || request.URL.User != nil ||
		request.URL.Path == "" || request.URL.Path[0] != '/' ||
		request.URL.RawPath != "" || request.URL.Fragment != "" {
		return nil, NewFault(FaultInvalid)
	}
	target, exists := client.targets[domain]
	if !exists || target.target.domain != domain {
		return nil, NewFault(FaultUnauthorized)
	}
	wire, err := credential.Wire()
	if err != nil {
		return nil, NewFault(FaultInvalid)
	}
	prepared := request.Clone(request.Context())
	prepared.URL = &url.URL{Scheme: target.target.baseURL.Scheme,
		Host: target.target.baseURL.Host, Path: request.URL.Path,
		RawQuery: request.URL.RawQuery, ForceQuery: request.URL.ForceQuery}
	prepared.Host = ""
	prepared.RequestURI = ""
	prepared.Header = request.Header.Clone()
	stripActorBridgeHeaders(prepared.Header)
	prepared.Header.Set(actorassertionv1.HeaderName, wire)
	prepared.TransferEncoding = nil
	prepared.Trailer = nil
	return prepared, nil
}
