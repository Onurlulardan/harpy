package application

import (
	"crypto/ed25519"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/google/uuid"
)

type DevelopmentBindingProofInput struct {
	EnvironmentID             string
	LicenseBindingID          string
	FinalizationReceiptSHA256 string
	RecoveryTrustBundleSHA256 string
	ControlKID                string
	OperationID               string
}

func SignDevelopmentBindingProof(
	input DevelopmentBindingProofInput,
	private ed25519.PrivateKey,
) ([]byte, error) {
	operationID, err := uuid.Parse(input.OperationID)
	if err != nil {
		return nil, fmt.Errorf("development authority operation is invalid")
	}
	nonce := uuid.NewSHA1(operationID, []byte("development-registry-nonce"))
	proof, err := artifact.SignBindingProof(artifact.BindingProof{
		Schema: artifact.BindingSchema, EnvironmentID: input.EnvironmentID,
		LicenseBindingID:          input.LicenseBindingID,
		FinalizationReceiptSHA256: input.FinalizationReceiptSHA256,
		RecoveryTrustBundleSHA256: input.RecoveryTrustBundleSHA256,
		ControlKID:                input.ControlKID, RegistryNonce: nonce.String(),
		OperationID: input.OperationID,
	}, private)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), proof.Bytes...), nil
}
