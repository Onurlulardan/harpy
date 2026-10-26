package recoveryplansigner

import (
	"context"
	"crypto/ed25519"
	"fmt"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Signer struct {
	bundle  recoverysigningv1.Bundle
	private ed25519.PrivateKey
}

func New(bundleBytes, controlSeed []byte) (*Signer, error) {
	bundle, err := recoverysigningv1.Parse(bundleBytes)
	if err != nil || len(controlSeed) != ed25519.SeedSize {
		return nil, fmt.Errorf("restore plan signer input is invalid")
	}
	private := ed25519.NewKeyFromSeed(controlSeed)
	public, found := bundle.PublicKey(recoverysigningv1.PurposeRecoveryControl)
	if !found || !private.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(public[:])) {
		clear(private)
		return nil, fmt.Errorf("restore plan signer key does not match trust bundle")
	}
	return &Signer{bundle: bundle, private: private}, nil
}

func (signer *Signer) SignRestorePlan(
	ctx context.Context,
	request ports.RestorePlanSigningRequest,
) (ports.RestorePlanSigningResult, error) {
	if signer == nil || ctx == nil || ctx.Err() != nil || request.EnvironmentID == "" ||
		request.OperationID == "" || request.Sequence < 1 || len(request.Payload) == 0 ||
		request.EnvironmentID != signer.bundle.EnvironmentID() {
		return ports.RestorePlanSigningResult{}, fmt.Errorf("restore plan signing request is invalid")
	}
	artifact, err := environmentrecoveryv1.Build(environmentrecoveryv1.Input{
		Purpose:       environmentrecoveryv1.PurposeRestorePlan,
		EnvironmentID: request.EnvironmentID, OperationID: request.OperationID,
		ArtifactID: request.OperationID, Sequence: request.Sequence,
		IssuedAt: request.IssuedAt, ExpiresAt: request.ExpiresAt,
		Payload: request.Payload,
	}, signer.bundle, signer.private)
	if err != nil {
		return ports.RestorePlanSigningResult{}, fmt.Errorf("sign restore plan: %w", err)
	}
	return ports.RestorePlanSigningResult{Generation: signer.bundle.Generation(),
		Document: artifact.CanonicalBytes()}, nil
}

var _ ports.RestorePlanSigner = (*Signer)(nil)
