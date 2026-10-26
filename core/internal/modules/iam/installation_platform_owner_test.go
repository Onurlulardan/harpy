package iam

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/google/uuid"
)

type installationPlatformOwnerTestRow func(...any) error

func (row installationPlatformOwnerTestRow) Scan(destinations ...any) error {
	return row(destinations...)
}

func TestInitializeInstallationPlatformOwnerKeepsBindV2(t *testing.T) {
	input := installationPlatformOwnerTestInput()
	calls := 0
	executor := InstallationActivationExecutor{QueryRow: func(
		_ context.Context, statement string, arguments ...any,
	) InstallationActivationRow {
		calls++
		if !strings.Contains(statement, "iam_installation_instance_owner_bind_v2") {
			t.Fatalf("initialize query = %q", statement)
		}
		assertInstallationPlatformOwnerArguments(t, arguments,
			installationPlatformOwnerRecord(input), true)
		return installationPlatformOwnerBooleanRow(false)
	}}
	if err := InitializeInstallationPlatformOwner(
		context.Background(), executor, input,
	); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("initialize calls = %d", calls)
	}
}

func TestReconcileInstallationPlatformOwnerUsesVerifier(t *testing.T) {
	input := installationPlatformOwnerTestInput()
	executor, calls := installationPlatformOwnerReconcileExecutor(t, input, true)
	if err := ReconcileInstallationPlatformOwner(
		context.Background(), executor, input.EnvironmentID,
	); err != nil {
		t.Fatal(err)
	}
	if *calls != 2 {
		t.Fatalf("reconcile calls = %d", *calls)
	}
}

func TestReconcileInstallationPlatformOwnerFailsClosedOnFalse(t *testing.T) {
	input := installationPlatformOwnerTestInput()
	executor, _ := installationPlatformOwnerReconcileExecutor(t, input, false)
	err := ReconcileInstallationPlatformOwner(
		context.Background(), executor, input.EnvironmentID,
	)
	if !errors.Is(err, ErrInstallationPlatformOwnerUnverified) {
		t.Fatalf("false verification error = %v", err)
	}
}

func installationPlatformOwnerReconcileExecutor(
	t *testing.T,
	input InstallationPlatformOwnerInput,
	verified bool,
) (InstallationActivationExecutor, *int) {
	t.Helper()
	calls := 0
	executor := InstallationActivationExecutor{QueryRow: func(
		_ context.Context, statement string, arguments ...any,
	) InstallationActivationRow {
		calls++
		switch calls {
		case 1:
			if !strings.Contains(statement, "iam_installation_instance_owner_input_v1") {
				t.Fatalf("reconcile input query = %q", statement)
			}
			if !reflect.DeepEqual(arguments, []any{uuid.UUID(input.EnvironmentID)}) {
				t.Fatalf("reconcile input arguments = %#v", arguments)
			}
			return installationPlatformOwnerInputRow(input)
		case 2:
			if !strings.Contains(statement, "iam_installation_instance_owner_verify_v1") {
				t.Fatalf("reconcile verifier query = %q", statement)
			}
			assertInstallationPlatformOwnerArguments(t, arguments,
				installationPlatformOwnerRecord(input), false)
			return installationPlatformOwnerBooleanRow(verified)
		default:
			t.Fatalf("unexpected reconcile query = %q", statement)
			return nil
		}
	}}
	return executor, &calls
}

func assertInstallationPlatformOwnerArguments(
	t *testing.T,
	actual []any,
	record iampostgres.InstallationPlatformOwnerRecord,
	includeSchema bool,
) {
	t.Helper()
	expected := []any{
		record.OperationID, record.EnvironmentID, record.TenantID,
	}
	if includeSchema {
		expected = append(expected, record.SchemaName)
	}
	expected = append(expected,
		record.UserID, record.SubjectID, record.OwnerRoleID,
		record.RolePermissionID, record.OwnerBindingID,
		record.AuthorizationVersion, record.RequestSHA256,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("platform owner arguments:\nactual: %#v\nwant: %#v", actual, expected)
	}
}

func installationPlatformOwnerBooleanRow(value bool) InstallationActivationRow {
	return installationPlatformOwnerTestRow(func(destinations ...any) error {
		*destinations[0].(*bool) = value
		return nil
	})
}

func installationPlatformOwnerInputRow(
	input InstallationPlatformOwnerInput,
) InstallationActivationRow {
	return installationPlatformOwnerTestRow(func(destinations ...any) error {
		*destinations[0].(*uuid.UUID) = uuid.UUID(input.OperationID)
		*destinations[1].(*uuid.UUID) = uuid.UUID(input.TenantID)
		*destinations[2].(*string) = input.SchemaName
		*destinations[3].(*uuid.UUID) = uuid.UUID(input.FirstUserID)
		*destinations[4].(*uuid.UUID) = uuid.UUID(input.InstanceOwnerBindingID)
		*destinations[5].(*string) = input.RequestSHA256
		return nil
	})
}

func installationPlatformOwnerTestInput() InstallationPlatformOwnerInput {
	return InstallationPlatformOwnerInput{
		OperationID:   uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		EnvironmentID: uuid.MustParse("10000000-0000-4000-8000-000000000002"),
		TenantID:      uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		FirstUserID:   uuid.MustParse("10000000-0000-4000-8000-000000000004"),
		InstanceOwnerBindingID: uuid.MustParse(
			"10000000-0000-4000-8000-000000000005",
		),
		SchemaName:    "t_harpy",
		RequestSHA256: strings.Repeat("a", 64),
	}
}
