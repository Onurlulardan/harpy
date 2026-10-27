package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type authorizationQueryStub struct {
	row       platformpostgres.Row
	statement string
	arguments []any
}

func (stub *authorizationQueryStub) QueryRow(
	_ context.Context,
	statement string,
	arguments ...any,
) platformpostgres.Row {
	stub.statement = statement
	stub.arguments = append([]any(nil), arguments...)
	return stub.row
}

type authorizationRowStub struct {
	documents [5][]byte
	clock     int64
	err       error
}

func (row authorizationRowStub) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 6 {
		return errors.New("unexpected authorization destination count")
	}
	for index, source := range row.documents[:3] {
		*destinations[index].(*[]byte) = append([]byte(nil), source...)
	}
	*destinations[3].(*int64) = row.clock
	*destinations[4].(*[]byte) = append([]byte(nil), row.documents[3]...)
	*destinations[5].(*[]byte) = append([]byte(nil), row.documents[4]...)
	return nil
}

func TestReadEvidenceUsesOneTypedStatement(t *testing.T) {
	actorID := uuid.MustParse("11000000-0000-4000-8000-000000000001")
	stub := &authorizationQueryStub{row: authorizationRowStub{
		documents: [5][]byte{
			[]byte(`{"id":"11000000-0000-4000-8000-000000000001","status":"active","token_version":7,"authorization_version":9}`),
			[]byte(`{"id":"12000000-0000-4000-8000-000000000001","state":"active","protected":true}`),
			[]byte(`{"id":"13000000-0000-4000-8000-000000000001","state":"active","protected":false}`),
			[]byte(`[{
                    "path":"user","binding_id":"15000000-0000-4000-8000-000000000001",
                    "binding_state":"active","permission_id":"14000000-0000-4000-8000-000000000001",
                    "permission_match_kind":"exact","permission_state":"active","permission_protected":false,
                    "permission_resource":"12000000-0000-4000-8000-000000000001",
                    "permission_action":"13000000-0000-4000-8000-000000000001",
                    "resource_state":"active","action_state":"active"
                }]`),
			[]byte(`[]`),
		},
		clock: 11,
	}}
	request, err := authorization.NewRequest("iam.users", "read", 7, 9)
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	snapshot, err := (&boundAuthorizer{database: stub, query: authorizeSQL}).readEvidence(
		context.Background(), tenantID, actorID, request,
	)
	if err != nil || snapshot.Validate() != nil {
		t.Fatalf("read evidence: snapshot=%v err=%v", snapshot.Validate(), err)
	}
	if stub.statement != authorizeSQL || len(stub.arguments) != 3 ||
		stub.arguments[0] != actorID || stub.arguments[1] != "iam.users" ||
		stub.arguments[2] != "read" {
		t.Fatalf("unexpected query invocation: %#v", stub.arguments)
	}
}

func TestReadEvidenceFailsClosedOnDatabaseError(t *testing.T) {
	request, _ := authorization.NewRequest("iam.users", "read", 1, 1)
	stub := &authorizationQueryStub{row: authorizationRowStub{err: errors.New("offline")}}
	if _, err := (&boundAuthorizer{database: stub, query: authorizeSQL}).readEvidence(
		context.Background(), uuid.New(), uuid.New(), request,
	); !errors.Is(err, errAuthorizationRead) || !strings.Contains(err.Error(), "offline") {
		t.Fatalf("error = %v", err)
	}
}

func TestFormsReadEvidencePassesTenantToV2Routine(t *testing.T) {
	tenantID, actorID := uuid.New(), uuid.New()
	stub := &authorizationQueryStub{row: authorizationRowStub{err: errors.New("stop")}}
	request, _ := authorization.NewRequest("forms.records", "create", 1, 1)
	_, _ = (&boundAuthorizer{
		database: stub, query: authorizeFormsSQL, includesTenant: true,
	}).readEvidence(context.Background(), tenantID, actorID, request)
	if stub.statement != authorizeFormsSQL || len(stub.arguments) != 4 ||
		stub.arguments[0] != tenantID || stub.arguments[1] != actorID ||
		stub.arguments[2] != "forms.records" || stub.arguments[3] != "create" {
		t.Fatalf("unexpected Forms evidence invocation: %#v", stub.arguments)
	}
}
