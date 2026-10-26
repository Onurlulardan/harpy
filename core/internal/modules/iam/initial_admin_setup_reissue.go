package iam

import (
	"context"
	"fmt"
	"regexp"
	"time"

	setuppostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/initialadminsetup"
	setupapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/initialadminsetup"
	"github.com/google/uuid"
)

var initialAdminSetupReissueSchema = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type InitialAdminSetupReissueInput struct {
	OperationID, ReissueOperationID, TenantID, UserID uuid.UUID
	SchemaName, Email                                 string
	CreatedAt                                         time.Time
}

type InitialAdminSetupReissuer struct {
	pool     setuppostgres.RuntimePool
	preparer *setupapp.Preparer
}

func NewInitialAdminSetupReissuer(
	pool setuppostgres.RuntimePool,
	preparer *setupapp.Preparer,
) (*InitialAdminSetupReissuer, error) {
	if pool == nil || preparer == nil {
		return nil, fmt.Errorf("initial admin setup reissuer dependencies are required")
	}
	return &InitialAdminSetupReissuer{pool: pool, preparer: preparer}, nil
}

func (reissuer *InitialAdminSetupReissuer) Reissue(
	ctx context.Context,
	input InitialAdminSetupReissueInput,
) error {
	if reissuer == nil || reissuer.pool == nil || reissuer.preparer == nil ||
		ctx == nil || ctx.Err() != nil || input.OperationID == uuid.Nil ||
		input.ReissueOperationID == uuid.Nil || input.TenantID == uuid.Nil ||
		input.UserID == uuid.Nil || input.OperationID == input.ReissueOperationID ||
		!initialAdminSetupReissueSchema.MatchString(input.SchemaName) {
		return fmt.Errorf("initial admin setup reissue input is invalid")
	}
	stage, err := reissuer.preparer.BuildStage(ctx, setupapp.PrepareInput{
		OperationID: input.ReissueOperationID, TenantID: input.TenantID,
		UserID:       input.UserID,
		IdentifierID: InstallationFirstEmailIdentifierID(input.OperationID),
		Email:        input.Email, CreatedAt: input.CreatedAt,
	})
	if err != nil {
		return err
	}
	return setuppostgres.RunReissue(
		ctx, reissuer.pool, input.SchemaName,
		func(queryRow setuppostgres.QueryRow) error {
			store, storeErr := setuppostgres.NewStore(queryRow)
			if storeErr != nil {
				return storeErr
			}
			if storeErr = store.Stage(ctx, stage); storeErr != nil {
				return storeErr
			}
			installation, installErr := setuppostgres.NewInstallation(queryRow)
			if installErr != nil {
				return installErr
			}
			artifactID, installErr := installation.Reissue(ctx, setuppostgres.ReissueInput{
				OperationID:        input.OperationID,
				ReissueOperationID: input.ReissueOperationID,
				TenantID:           input.TenantID,
			})
			if installErr != nil {
				return fmt.Errorf("reissue initial admin setup: %w", installErr)
			}
			if artifactID != stage.Envelope.ArtifactID {
				return fmt.Errorf("reissued initial admin setup artifact is invalid")
			}
			return nil
		},
	)
}
