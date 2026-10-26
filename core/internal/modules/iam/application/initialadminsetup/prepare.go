package initialadminsetup

import (
	"context"
	"crypto/sha256"
	"fmt"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/lifecyclecredential"
	"github.com/google/uuid"
)

type Preparer struct{ dependencies Dependencies }

func NewPreparer(dependencies Dependencies) (*Preparer, error) {
	if dependencies.Store == nil || dependencies.Keys == nil ||
		dependencies.Delivery == nil {
		return nil, ErrInvalid
	}
	return &Preparer{dependencies: dependencies}, nil
}

func (preparer *Preparer) Prepare(
	ctx context.Context,
	input PrepareInput,
) (uuid.UUID, error) {
	stage, err := preparer.BuildStage(ctx, input)
	if err != nil {
		return uuid.Nil, err
	}
	if err := preparer.dependencies.Store.Stage(ctx, stage); err != nil {
		return uuid.Nil, fmt.Errorf("persist initial admin setup stage: %w", err)
	}
	return stage.Envelope.ArtifactID, nil
}

func (preparer *Preparer) BuildStage(
	ctx context.Context,
	input PrepareInput,
) (Stage, error) {
	if preparer == nil || ctx == nil || ctx.Err() != nil || !input.valid() {
		return Stage{}, ErrInvalid
	}
	tokenID := childID(input.OperationID, "initial-admin-setup-token")
	artifactID := childID(input.OperationID, "initial-admin-setup-artifact")
	issued, token, err := lifecyclecredential.NewGenerator().Generate(
		input.TenantID, lifecyclecredential.PurposeInitialAdminSetup, tokenID,
		preparer.dependencies.Keys,
	)
	if err != nil {
		return Stage{}, ErrUnavailable
	}
	defer issued.Destroy()
	var envelope lifecycledeliveryv1.Envelope
	err = issued.RevealOnce(func(plaintext []byte) error {
		var sealErr error
		envelope, sealErr = preparer.dependencies.Delivery.SealLifecycleCredential(
			ctx, input.TenantID, artifactID, input.UserID, input.IdentifierID,
			tokenID, lifecycledeliveryv1.PurposeInitialAdminSetup, plaintext,
		)
		return sealErr
	})
	reference, referenceErr := lifecycledeliveryv1.NewReference(input.TenantID, artifactID)
	stage := Stage{PrepareInput: input, Token: token, Envelope: envelope}
	if referenceErr == nil {
		stage.Reference = reference.String()
	}
	if err != nil || referenceErr != nil || stage.Validate() != nil {
		return Stage{}, ErrUnavailable
	}
	return stage, nil
}

func childID(operationID uuid.UUID, label string) uuid.UUID {
	digest := sha256.New()
	_, _ = digest.Write([]byte("harpy/iam/initial-admin-setup/v1\x00"))
	_, _ = digest.Write(operationID[:])
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(label))
	value := digest.Sum(nil)
	result, _ := uuid.FromBytes(value[:16])
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result
}
