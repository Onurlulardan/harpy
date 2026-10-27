package postgres

import (
	"context"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/google/uuid"
)

func TestActorBridgeIssueReadsExactAuthority(t *testing.T) {
	queryer := &actorBridgeTestQueryer{row: actorBridgeTestRow{scan: func(values ...any) error {
		*values[0].(*int64), *values[1].(*int64), *values[2].(*int64) = 5, 6, 7
		return nil
	}}}
	database, err := NewActorBridgeDatabase(queryer)
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	authority, err := database.ResolveIssueAuthority(context.Background(), tenantID,
		actorassertionv1.DomainDocuments)
	if err != nil {
		t.Fatal(err)
	}
	if authority.IAMGroupGeneration() != 5 || authority.IAMGeneration() != 6 ||
		authority.TargetGeneration() != 7 || queryer.query != issueIAMV1ActorAssertionSQL ||
		len(queryer.arguments) != 2 || queryer.arguments[0] != tenantID ||
		queryer.arguments[1] != "forms" {
		t.Fatal("actor bridge issue authority query differs")
	}
}
