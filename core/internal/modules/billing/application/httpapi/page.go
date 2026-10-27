package httpapi

import "encoding/json"

import restqueryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

type QueryRequest = restqueryapi.QueryRequest

type Query struct {
	Access    Access
	Request   restqueryapi.QueryRequest
	Transport string
}

type PublicQuery struct {
	Request   restqueryapi.QueryRequest
	Transport string
}

type Pagination struct {
	Next          *string `json:"next"`
	NextTransport string  `json:"next_transport"`
	Size          int     `json:"size"`
	Total         *int64  `json:"total,omitempty"`
}

type Page[T any] struct {
	Items      []T            `json:"items"`
	Pagination Pagination     `json:"pagination"`
	Meta       map[string]any `json:"meta"`
}

type QueryPage struct {
	Items      []json.RawMessage `json:"items"`
	Pagination Pagination        `json:"pagination"`
	Meta       map[string]any    `json:"meta"`
}
