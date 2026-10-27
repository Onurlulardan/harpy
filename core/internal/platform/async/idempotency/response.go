package idempotency

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

const maximumResponseBodyBytes = 64 * 1024

var allowedResponseHeaders = map[string]struct{}{
	"cache-control":    {},
	"content-language": {},
	"content-type":     {},
	"etag":             {},
	"last-modified":    {},
	"retry-after":      {},
}

type Response struct {
	status          int
	headers         map[string]string
	body            []byte
	resultReference *string
	sha256          string
}

func NewBodyResponse(
	status int,
	headers map[string]string,
	body []byte,
) (Response, error) {
	if body == nil {
		body = []byte{}
	}
	if len(body) > maximumResponseBodyBytes {
		return Response{}, ErrInvalidResponse
	}
	clone := make([]byte, len(body))
	copy(clone, body)
	return newResponse(status, headers, clone, nil)
}

func NewReferenceResponse(
	status int,
	headers map[string]string,
	resultReference string,
) (Response, error) {
	if len(resultReference) < 1 || len([]byte(resultReference)) > 2048 ||
		!utf8.ValidString(resultReference) || strings.ContainsAny(resultReference, "\r\n\x00") {
		return Response{}, ErrInvalidResponse
	}
	value := resultReference
	return newResponse(status, headers, nil, &value)
}

func newResponse(
	status int,
	headers map[string]string,
	body []byte,
	resultReference *string,
) (Response, error) {
	normalized, err := normalizeHeaders(headers)
	if err != nil || status < 100 || status > 599 || (body == nil) == (resultReference == nil) {
		return Response{}, ErrInvalidResponse
	}
	value := Response{
		status: status, headers: normalized, body: body,
		resultReference: resultReference,
	}
	value.sha256 = digestResponse(value)
	return value, nil
}

func (response Response) Validate() error {
	if response.status < 100 || response.status > 599 ||
		(response.body == nil) == (response.resultReference == nil) ||
		!sha256Pattern.MatchString(response.sha256) ||
		digestResponse(response) != response.sha256 {
		return ErrInvalidResponse
	}
	_, err := normalizeHeaders(response.headers)
	if err != nil || response.body != nil && len(response.body) > maximumResponseBodyBytes {
		return ErrInvalidResponse
	}
	if response.resultReference != nil &&
		(len(*response.resultReference) < 1 || len([]byte(*response.resultReference)) > 2048 ||
			!utf8.ValidString(*response.resultReference) || strings.ContainsAny(*response.resultReference, "\r\n\x00")) {
		return ErrInvalidResponse
	}
	return nil
}

func normalizeHeaders(headers map[string]string) (map[string]string, error) {
	if len(headers) > 16 {
		return nil, ErrInvalidResponse
	}
	normalized := make(map[string]string, len(headers))
	for rawName, value := range headers {
		name := strings.ToLower(strings.TrimSpace(rawName))
		if _, exists := allowedResponseHeaders[name]; !exists ||
			len(value) > 1024 || !utf8.ValidString(value) || strings.ContainsAny(value, "\r\n\x00") {
			return nil, ErrInvalidResponse
		}
		if _, duplicate := normalized[name]; duplicate {
			return nil, ErrInvalidResponse
		}
		normalized[name] = value
	}
	encoded, err := json.Marshal(normalized)
	if err != nil || len(encoded)+len(normalized)*2 > 16384 {
		return nil, ErrInvalidResponse
	}
	return normalized, nil
}

func (response Response) Status() int    { return response.status }
func (response Response) SHA256() string { return response.sha256 }

func (response Response) Headers() map[string]string {
	clone := make(map[string]string, len(response.headers))
	for name, value := range response.headers {
		clone[name] = value
	}
	return clone
}

func (response Response) HeadersJSON() []byte {
	encoded, _ := json.Marshal(response.headers)
	return encoded
}

func (response Response) Body() []byte {
	if response.body == nil {
		return nil
	}
	clone := make([]byte, len(response.body))
	copy(clone, response.body)
	return clone
}

func (response Response) ResultReference() *string {
	if response.resultReference == nil {
		return nil
	}
	value := *response.resultReference
	return &value
}
