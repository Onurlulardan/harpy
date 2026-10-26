package api

import (
	"context"
	"io"
)

type Expected struct {
	Identity         Identity
	CiphertextBytes  int64
	CiphertextSHA256 [32]byte
	HeaderSHA256     [32]byte
}

type Reader interface {
	OpenExact(context.Context, Expected) (io.ReadCloser, error)
}
