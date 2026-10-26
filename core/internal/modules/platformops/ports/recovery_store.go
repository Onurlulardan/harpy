package ports

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("platform operations row was not found")
	ErrConflict = errors.New("platform operations row conflicts")
)

type SigningBundleRecord struct {
	EnvironmentID string
	Generation    int64
	SHA256        string
	BundleBytes   []byte
	CreatedAt     time.Time
}

type RecoveryStore interface {
	RecordSigningBundle(context.Context, SigningBundleRecord) error
	LoadSigningBundle(context.Context, string, int64) (SigningBundleRecord, error)
}
