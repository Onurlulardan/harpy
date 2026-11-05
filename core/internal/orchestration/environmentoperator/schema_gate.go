package environmentoperator

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

var operatorSchemaPattern = regexp.MustCompile(`^tenant_[0-9a-f]{32}$`)

type tenantReadiness interface {
	SchemaReady(context.Context, uuid.UUID) error
	TenantActivated(context.Context, uuid.UUID) (bool, error)
}

func (coordinator *Coordinator) prepareSchema(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
) error {
	if err := validPlanSchema(plan); err != nil {
		return err
	}
	ready, ok := coordinator.provision.(tenantReadiness)
	if !ok {
		return nil
	}
	err := ready.SchemaReady(ctx, uuid.UUID(plan.TenantID()))
	if err == nil || errorsIsTerminal(err) {
		return err
	}
	return Resumable("tenant_schema.provision", "provision_failed", "", err)
}

func (coordinator *Coordinator) tenantActivated(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
) (bool, error) {
	ready, ok := coordinator.provision.(tenantReadiness)
	if !ok {
		return true, nil
	}
	return ready.TenantActivated(ctx, uuid.UUID(plan.OperationID()))
}

func validPlanSchema(plan tenantapi.OnboardingPlan) error {
	want, err := schemaNameForTenant(uuid.UUID(plan.TenantID()))
	got := plan.SchemaName()
	if err != nil || got != want || !operatorSchemaPattern.MatchString(got) {
		return invalidTenantSchema(got)
	}
	return nil
}

func schemaNameForTenant(tenantID uuid.UUID) (string, error) {
	if tenantID == uuid.Nil {
		return "", invalidTenantSchema("")
	}
	name := "tenant_" + hex.EncodeToString(tenantID[:])
	if !operatorSchemaPattern.MatchString(name) {
		return "", invalidTenantSchema(name)
	}
	return name, nil
}

func invalidTenantSchema(name string) error {
	if name == "" {
		name = "<empty>"
	}
	return fmt.Errorf("%w: tenant schema name is invalid: %s", ErrInvalid, name)
}

func errorsIsTerminal(err error) bool {
	return errors.Is(err, ErrInvalid) || errors.Is(err, ErrFailed) ||
		errors.Is(err, ErrFenced) || errors.Is(err, ErrDependencyMismatch)
}
