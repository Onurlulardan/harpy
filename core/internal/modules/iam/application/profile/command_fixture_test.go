package profile

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	fixtureActorID      = uuid.MustParse("51000000-0000-4000-8000-000000000001")
	fixturePermissionID = uuid.MustParse("51000000-0000-4000-8000-000000000002")
	fixtureOperationID  = uuid.MustParse("51000000-0000-4000-8000-000000000003")
	fixtureActiveSchema = uuid.MustParse("51000000-0000-4000-8000-000000000004")
	fixtureTargetSchema = uuid.MustParse("51000000-0000-4000-8000-000000000005")
	fixtureFieldID      = uuid.MustParse("51000000-0000-4000-8000-000000000006")
	fixtureResourceID   = uuid.MustParse("51000000-0000-4000-8000-000000000007")
)

func fixtureEvidence() MutationEvidence {
	return MutationEvidence{
		OperationID: fixtureOperationID, RequestSHA256: [32]byte{1},
		ActorID: fixtureActorID, ManagePermissionID: fixturePermissionID,
		ExpectedTokenVersion: 2, ExpectedAuthorizationVersion: 3,
	}
}

func fixtureCommandSchema(t *testing.T, state FieldState) Schema {
	t.Helper()
	field, err := NewField(FieldInput{
		ID: fixtureFieldID, Resource: ResourceUser, Key: "display_name",
		Type: LogicalString, Cardinality: CardinalitySingle,
		Classification: ClassificationInternal,
		Normalization:  NormalizationNoneV1, State: state,
		Validation: ValidationInput{MaximumBytes: 100},
	})
	if err != nil {
		t.Fatalf("create field: %v", err)
	}
	schema, err := NewPublishedSchema(SchemaInput{
		ID: fixtureTargetSchema, Resource: ResourceUser,
		Version: 2, State: SchemaActive, Fields: []Field{field},
	})
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return schema
}

type fixturePort struct{}

func (fixturePort) GetSelf(
	context.Context, transaction.TenantScope, SelfEvidence,
) (SelfRecord, error) {
	return SelfRecord{}, ErrEvidenceMissing
}
func (fixturePort) UpdateSelf(
	context.Context, transaction.TenantScope, SelfUpdateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}

func (fixturePort) Publish(
	context.Context, transaction.TenantScope, PublishCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (fixturePort) Backfill(
	context.Context, transaction.TenantScope, BackfillCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (fixturePort) Retire(
	context.Context, transaction.TenantScope, RetireCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}

type fixtureReferences struct{}

func (fixtureReferences) AssertRetirable(
	context.Context, transaction.TenantScope, RetirementTarget,
) (ReferenceEvidence, error) {
	return ReferenceEvidence{}, ErrEvidenceMissing
}

func fixtureCommands(t *testing.T) *Commands {
	t.Helper()
	commands, err := NewCommands(fixturePort{}, fixtureReferences{})
	if err != nil {
		t.Fatalf("create commands: %v", err)
	}
	return commands
}
