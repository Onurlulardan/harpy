package documentshttp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/url"
)

type Request struct {
	Access         Access
	Operation      string
	ResourceID     string
	SecondaryID    string
	IdempotencyKey string
	Query          url.Values
	Document       json.RawMessage
	Content        io.Reader
}

type Binary struct {
	Filename string
	MIME     string
	Bytes    int64
	SHA256   [sha256.Size]byte
	Content  io.ReadCloser
}

type Response struct {
	Status   int
	Data     any
	Location string
	Binary   *Binary
}

type Application interface {
	Dispatch(context.Context, Request) (Response, error)
}
