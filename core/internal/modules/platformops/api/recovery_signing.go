package api

import "context"

type RecoverySigningRequest struct {
	EnvironmentID string
	BundleBytes   []byte
}

type RecoverySigningBundle struct {
	EnvironmentID string
	Generation    int64
	SHA256        string
	BundleBytes   []byte
}

type RecoverySigning interface {
	RecordRecoverySigning(context.Context, RecoverySigningRequest) (RecoverySigningBundle, error)
	ExportRecoverySigning(context.Context, string, int64) (RecoverySigningBundle, error)
}
