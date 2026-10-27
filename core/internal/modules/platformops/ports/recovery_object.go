package ports

import (
	"io"
	"time"
)

type ObjectIdentity struct {
	Key       string
	VersionID string
	SHA256    string
	Bytes     int64
}

type WriteRequest struct {
	Key            string
	Body           io.Reader
	ContentLength  int64
	ExpectedSHA256 string
	ExpectedBytes  int64
}

type WriteEvidence struct {
	ObjectIdentity
	RetentionMode  string
	RetainUntil    time.Time
	UploadSHA256   string
	UploadedBytes  int64
	ReadbackSHA256 string
	ReadbackBytes  int64
}
