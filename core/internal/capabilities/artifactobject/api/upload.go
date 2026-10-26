package api

import (
	"context"
	"io"
	"time"
)

type Upload struct {
	Key              string
	Body             io.ReadSeeker
	CiphertextBytes  int64
	CiphertextSHA256 [32]byte
	HeaderSHA256     [32]byte
}

type UploadResult struct {
	Identity          Identity
	ProviderRequestID string
	UploadedAt        time.Time
}

type Uploader interface {
	Upload(context.Context, Upload) (UploadResult, error)
}
