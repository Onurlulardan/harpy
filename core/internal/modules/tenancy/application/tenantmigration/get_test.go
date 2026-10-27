package tenantmigration

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type getRepository struct{ job tenantapi.TenantMigrationJob }

func (repository getRepository) CurrentReady(
	context.Context,
) (tenantapi.TenantMigrationReady, bool, error) {
	return tenantapi.TenantMigrationReady{}, false, nil
}

func (repository getRepository) Enqueue(
	context.Context, EnqueueBinding,
) (tenantapi.TenantMigrationJob, error) {
	return tenantapi.TenantMigrationJob{}, nil
}

func (repository getRepository) Get(
	context.Context, uuid.UUID,
) (tenantapi.TenantMigrationJob, bool, error) {
	return repository.job, true, nil
}

func TestGetReturnsValidatedMigrationJob(t *testing.T) {
	operationID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	job := tenantapi.TenantMigrationJob{
		JobID:       "22222222-2222-4222-8222-222222222222",
		OperationID: operationID.String(), State: "running",
		TenantID:          "33333333-3333-4333-8333-333333333333",
		SchemaName:        "t_33333333333343338333333333333333",
		StepPlanSHA256:    strings.Repeat("a", 64),
		ReleaseID:         "44444444-4444-4444-8444-444444444444",
		RuntimeGeneration: 13, DatabaseGeneration: 15,
		MigrationManifestSHA256: strings.Repeat("b", 64),
		CreatedAt:               time.Date(2026, 7, 22, 11, 59, 0, 0, time.UTC),
		DatabaseNow:             time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
	}
	service, err := NewService(getRepository{job: job}, uuid.NewRandom)
	if err != nil {
		t.Fatal(err)
	}
	got, found, err := service.Get(context.Background(), operationID)
	if err != nil || !found || got.OperationID != operationID.String() {
		t.Fatalf("Get() = %#v, %v, %v", got, found, err)
	}
}
