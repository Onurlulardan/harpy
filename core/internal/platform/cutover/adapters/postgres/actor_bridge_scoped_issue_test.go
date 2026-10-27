package postgres

import (
	"context"
	"errors"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/google/uuid"
)

type actorBridgeScopedRow struct{ values [3]int64 }

func (row actorBridgeScopedRow) Scan(destinations ...any) error {
	if len(destinations) != 3 {
		return errors.New("unexpected scan")
	}
	for index, value := range row.values {
		*destinations[index].(*int64) = value
	}
	return nil
}

func TestActorBridgeIssueDatabaseUsesNarrowBoundQuery(t *testing.T) {
	tenantID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	var statement string
	var arguments []any
	database, err := NewActorBridgeIssueDatabase(func(
		_ context.Context,
		query string,
		values ...any,
	) ActorBridgeRow {
		statement = query
		arguments = append([]any(nil), values...)
		return actorBridgeScopedRow{values: [3]int64{11, 12, 13}}
	})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := database.ResolveIssueAuthority(
		context.Background(), tenantID, actorassertionv1.DomainDocuments,
	)
	if err != nil || authority.IAMGroupGeneration() != 11 ||
		authority.IAMGeneration() != 12 || authority.TargetGeneration() != 13 {
		t.Fatalf("authority=%v err=%v", authority, err)
	}
	if statement != issueIAMV1ActorAssertionSQL || len(arguments) != 2 ||
		arguments[0] != tenantID || arguments[1] != "forms" {
		t.Fatalf("query arguments = %#v", arguments)
	}
}
