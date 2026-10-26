package platformauthority

import (
	"context"
	"fmt"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	"github.com/google/uuid"
)

type SaaSBinding struct {
	BindingID, EnvironmentID, SignerKID, SignerPublicKeySHA256 string
	SignerGeneration                                           int64
}

func (binding SaaSBinding) Valid() bool {
	_, bindingErr := uuid.Parse(binding.BindingID)
	_, environmentErr := uuid.Parse(binding.EnvironmentID)
	return bindingErr == nil && environmentErr == nil &&
		binding.SignerGeneration > 0 && binding.SignerKID != "" &&
		len(binding.SignerPublicKeySHA256) == 64
}

func ResolveSaaSBinding(
	ctx context.Context,
	pool *postgresadapter.Pool,
	clock ports.Clock,
	environmentID string,
) (SaaSBinding, error) {
	if ctx == nil || ctx.Err() != nil || clock == nil {
		return SaaSBinding{}, fmt.Errorf("SaaS authority binding dependencies are invalid")
	}
	store, err := postgresadapter.NewStore(pool, clock)
	if err != nil {
		return SaaSBinding{}, err
	}
	record, err := store.LoadSaaSBindingForEnvironment(ctx, environmentID)
	if err != nil {
		return SaaSBinding{}, err
	}
	if record.DeploymentMode != "saas" ||
		record.RegistryKind != "saas_provider" || record.SignerStatus != "active" {
		return SaaSBinding{}, fmt.Errorf("active SaaS authority binding is unavailable")
	}
	result := SaaSBinding{BindingID: record.BindingID,
		EnvironmentID: record.EnvironmentID, SignerGeneration: record.SignerGeneration,
		SignerKID:             record.SignerKID,
		SignerPublicKeySHA256: record.SignerPublicKeySHA256}
	if !result.Valid() || result.EnvironmentID != environmentID {
		return SaaSBinding{}, fmt.Errorf("SaaS authority binding is invalid")
	}
	return result, nil
}
