package api

import (
	"context"
)

type OnPremiseBindingRequest struct {
	BindingID                 string
	EnvironmentID             string
	FinalizationReceiptSHA256 string
	RecoveryTrustBundleSHA256 string
	ProofBytes                []byte
}

type SaaSProviderBindingRequest struct {
	BindingID        string
	EnvironmentID    string
	SignerGeneration int64
}

type BindingReceipt struct {
	BindingID             string
	EnvironmentID         string
	RegistryKind          string
	SignerGeneration      int64
	SignerKID             string
	SignerPublicKeySHA256 string
}

type BindingBootstrap interface {
	BindOnPremise(context.Context, OnPremiseBindingRequest) (BindingReceipt, error)
	BindSaaSProvider(context.Context, SaaSProviderBindingRequest) (BindingReceipt, error)
}
