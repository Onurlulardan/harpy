package systemresourcequery

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
)

// Request is the IAM-owned opaque boundary around a validated REST Query request.
type Request struct {
	value     api.QueryRequest
	transport string
	valid     bool
}

func DecodeGET(rawQuery string) (Request, error) {
	value, err := httpdecode.GET().Decode(rawQuery)
	if err != nil {
		return Request{}, NormalizeError(err)
	}
	return newRequest(value, "GET")
}

func DecodePOST(document []byte) (Request, error) {
	value, err := httpdecode.POST().Decode(document)
	if err != nil {
		return Request{}, NormalizeError(err)
	}
	return newRequest(value, "POST")
}

func newRequest(value api.QueryRequest, transport string) (Request, error) {
	if value.Validate() != nil || transport != "GET" && transport != "POST" {
		return Request{}, ErrContract
	}
	return Request{value: value, transport: transport, valid: true}, nil
}

func (request Request) Value() (api.QueryRequest, error) {
	if !request.valid || request.value.Validate() != nil ||
		request.transport != "GET" && request.transport != "POST" {
		return api.QueryRequest{}, ErrContract
	}
	return request.value, nil
}

func (request Request) Transport() string {
	if _, err := request.Value(); err != nil {
		return ""
	}
	return request.transport
}
