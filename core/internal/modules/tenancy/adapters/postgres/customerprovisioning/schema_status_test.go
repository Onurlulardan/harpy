package customerprovisioning

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

var schemaStatusTenant = uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

const schemaStatusName = "tenant_00112233445566778899aabbccddeeff"

type scanned struct {
	scan func(...any) error
}

func (row scanned) Scan(dest ...any) error { return row.scan(dest...) }

func TestBeginReturnsRunErrorWhenIdentityIsAbsent(t *testing.T) {
	runErr := errors.New("tenant provisioning is waiting for host execution")
	var runs int
	status := newBeginStatus(t, scriptQuery(t, false), func(_ context.Context, tenantID string) error {
		runs++
		if tenantID != schemaStatusTenant.String() {
			t.Fatalf("tenant = %s", tenantID)
		}
		return runErr
	})
	err := status.Begin(context.Background(), schemaStatusTenant, schemaStatusName, "Acme", "acme")
	if !errors.Is(err, runErr) || runs != 1 {
		t.Fatalf("err=%v runs=%d", err, runs)
	}
}

func TestBeginSkipsRunWhenIdentityExists(t *testing.T) {
	var runs int
	status := newBeginStatus(t, scriptQuery(t, true), func(context.Context, string) error {
		runs++
		return errors.New("run")
	})
	err := status.Begin(context.Background(), schemaStatusTenant, schemaStatusName, "Acme", "acme")
	if err != nil || runs != 0 {
		t.Fatalf("err=%v runs=%d", err, runs)
	}
}

func TestBeginRejectsInvalidSchemaNameWithoutRun(t *testing.T) {
	var queries, runs int
	status := newBeginStatus(t, func(_ context.Context, statement string, args ...any) Row {
		queries++
		if queries != 1 || statement != markProvisioningSQL || len(args) != 4 {
			t.Fatalf("query %d statement=%q args=%v", queries, statement, args)
		}
		return scanned{scan: func(dest ...any) error {
			*(dest[0].(*uuid.UUID)) = schemaStatusTenant
			return nil
		}}
	}, func(context.Context, string) error {
		runs++
		return nil
	})
	err := status.Begin(context.Background(), schemaStatusTenant, "public", "Acme", "acme")
	if err == nil || runs != 0 || queries != 1 {
		t.Fatalf("err=%v runs=%d queries=%d", err, runs, queries)
	}
}

func TestBeginDoesNotRunWhenMarkProvisioningFails(t *testing.T) {
	markErr := errors.New("mark failed")
	var queries, runs int
	status := newBeginStatus(t, func(context.Context, string, ...any) Row {
		queries++
		return scanned{scan: func(...any) error { return markErr }}
	}, func(context.Context, string) error {
		runs++
		return nil
	})
	err := status.Begin(context.Background(), schemaStatusTenant, schemaStatusName, "Acme", "acme")
	if !errors.Is(err, markErr) || runs != 0 || queries != 1 {
		t.Fatalf("err=%v runs=%d queries=%d", err, runs, queries)
	}
}

func newBeginStatus(
	t *testing.T, query QueryRow, run func(context.Context, string) error,
) *SchemaStatus {
	t.Helper()
	status, err := NewSchemaStatus(query, run)
	if err != nil {
		t.Fatal(err)
	}
	return status
}

func scriptQuery(t *testing.T, identityExists bool) QueryRow {
	t.Helper()
	var calls int
	return func(_ context.Context, statement string, args ...any) Row {
		calls++
		switch calls {
		case 1:
			if statement != markProvisioningSQL || len(args) != 4 || args[0] != schemaStatusTenant ||
				args[3] != schemaStatusName {
				t.Fatalf("mark statement=%q args=%v", statement, args)
			}
			return scanned{scan: func(dest ...any) error {
				*(dest[0].(*uuid.UUID)) = schemaStatusTenant
				return nil
			}}
		case 2:
			if statement != tenantSchemaIdentitySQL || len(args) != 1 || args[0] != schemaStatusName {
				t.Fatalf("ready statement=%q args=%v", statement, args)
			}
			return scanned{scan: func(dest ...any) error {
				*(dest[0].(*bool)) = identityExists
				return nil
			}}
		default:
			t.Fatalf("unexpected query %d", calls)
			return scanned{}
		}
	}
}
