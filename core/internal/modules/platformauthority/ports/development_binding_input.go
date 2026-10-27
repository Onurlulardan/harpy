package ports

import "context"

type DevelopmentBindingInput struct {
	OperationID               string
	RegistryBindingID         string
	FinalizationReceiptSHA256 string
	RecoveryTrustBundleSHA256 string
}

type DevelopmentBindingInputReader interface {
	ReadDevelopmentBindingInput(
		context.Context, string,
	) (DevelopmentBindingInput, error)
}
