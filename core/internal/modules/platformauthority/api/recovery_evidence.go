package api

import (
	"context"
	"time"
)

type RecoveryHead struct {
	BindingID                 string
	EnvironmentID             string
	DeploymentMode            string
	RegistryKind              string
	LicenseBindingID          string
	FinalizationReceiptSHA256 string
	RecoveryTrustBundleSHA256 string
	SignerGeneration          int64
	SignerKID                 string
	SignerPublicKeySHA256     string
	BindingProofSHA256        string
	BindingProofBytes         []byte
	OperationID               string
	Purpose                   string
	ScopeKind                 string
	ScopeID                   string
	Generation                int64
	PreviousHeadSHA256        string
	StateSHA256               string
	StateBytes                []byte
	ProposalSHA256            string
	ProposalBytes             []byte
	AckSHA256                 string
	AckBytes                  []byte
	RegistrySequence          int64
	CompletedAt               time.Time
}

type RecoveryEvidence interface {
	ReadRecoveryHead(context.Context, string) (RecoveryHead, error)
}
