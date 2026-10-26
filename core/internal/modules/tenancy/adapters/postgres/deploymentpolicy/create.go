package deploymentpolicy

import (
	"context"
	"errors"
	"fmt"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type sqlStateError interface{ SQLState() string }

func (initializer *policyInitializer) Create(
	ctx context.Context,
	request tenantapi.InitializeDeploymentPolicyRequest,
) (tenantapi.DeploymentPolicyInitialization, error) {
	if !initializer.usable(ctx) || !request.IsValid() {
		return tenantapi.DeploymentPolicyInitialization{}, errPolicyUnavailable
	}
	var operationID, environmentID uuid.UUID
	var mode string
	var maximum *int64
	var createdAt, databaseNow time.Time
	var replayed bool
	err := initializer.queryRow(ctx, initializeDeploymentPolicySQL,
		uuid.UUID(request.OperationID), uuid.UUID(request.EnvironmentID),
		string(request.OperationKind), string(request.DeploymentMode)).Scan(
		&operationID, &environmentID, &mode, &maximum, &createdAt, &databaseNow, &replayed)
	if err != nil {
		var state sqlStateError
		if errors.As(err, &state) && state.SQLState() == "23505" {
			return tenantapi.DeploymentPolicyInitialization{},
				tenantapi.ErrDeploymentPolicyConflict
		}
		return tenantapi.DeploymentPolicyInitialization{},
			fmt.Errorf("execute deployment policy creation: %w", err)
	}
	policy, err := tenantapi.NewDeploymentPolicy(
		environmentID.String(), tenantapi.DeploymentMode(mode), maximum, createdAt.UTC())
	if err != nil {
		return tenantapi.DeploymentPolicyInitialization{}, errPolicyUnavailable
	}
	return tenantapi.DeploymentPolicyInitialization{
		OperationID: [16]byte(operationID), Policy: policy,
		DatabaseNow: databaseNow.UTC(), Replayed: replayed,
	}, nil
}
