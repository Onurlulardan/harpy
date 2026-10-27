package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestExactCountStatementSealsClassAndProvenance(t *testing.T) {
	statement := fixtureExactCountStatement(t)
	class, classErr := statement.Class()
	tenant, tenantErr := statement.TenantID()
	actor, actorErr := statement.ActorID()
	authorization, authorizationErr := statement.AuthorizationContextSHA256()
	mandatory, mandatoryErr := statement.MandatorySHA256()
	identity, identityErr := statement.QueryIdentitySHA256()
	continuation, continuationErr := statement.ContinuationSHA256()
	owner, ownerErr := statement.Owner()
	if classErr != nil || tenantErr != nil || actorErr != nil ||
		authorizationErr != nil || mandatoryErr != nil || identityErr != nil ||
		continuationErr != nil || ownerErr != nil ||
		class != planner.StatementClassExactCount || owner != "forms" ||
		tenant.String() != fixtureTenantID || actor.String() != fixtureActorID ||
		authorization == [32]byte{} || mandatory == [32]byte{} ||
		identity == [32]byte{} || continuation == [32]byte{} {
		t.Fatal("exact-count provenance did not round trip")
	}
}

func TestExactCountStatementRejectsTamperAndResealedWrongClass(t *testing.T) {
	statement := fixtureExactCountStatement(t)
	for name, mutate := range map[string]func(*ExactCountStatement){
		"sql":       func(value *ExactCountStatement) { value.sql += " OR TRUE" },
		"argument":  func(value *ExactCountStatement) { value.arguments[0] = "changed" },
		"binding":   func(value *ExactCountStatement) { value.binding.subject[0] ^= 0xff },
		"seal":      func(value *ExactCountStatement) { value.stateSeal[0] ^= 0xff },
		"validated": func(value *ExactCountStatement) { value.validated = false },
	} {
		t.Run(name, func(t *testing.T) {
			tampered := statement
			tampered.arguments = cloneArguments(statement.arguments)
			mutate(&tampered)
			if !errors.Is(tampered.Validate(), ErrContract) {
				t.Fatal("tampered exact-count statement was accepted")
			}
		})
	}
	tampered := statement
	tampered.class = planner.StatementClassRecord
	tampered.stateSeal, _ = exactCountStatementSeal(tampered)
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("resealed record class was accepted as exact count")
	}
	_, err := newExactCountStatement(
		statement.sql, statement.arguments,
		planner.StatementClassRecord, statement.binding,
	)
	if !errors.Is(err, ErrContract) {
		t.Fatal("wrong exact-count constructor class was accepted")
	}
}

func TestExactCountStatementDefensiveCopiesAndRedacts(t *testing.T) {
	base := fixtureExactCountStatement(t)
	values := []string{"OPEN", "CLOSED"}
	statement, err := newExactCountStatement(
		"SELECT COUNT(*) WHERE $1 IS NOT NULL", []any{values},
		planner.StatementClassExactCount, base.binding,
	)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = "MUTATED"
	arguments, _ := statement.Arguments()
	arguments[0].([]string)[0] = "RETURNED_MUTATION"
	again, _ := statement.Arguments()
	if again[0].([]string)[0] != "OPEN" || statement.Validate() != nil {
		t.Fatal("exact-count argument ownership escaped")
	}
	if fmt.Sprint(statement) != compilerRedaction {
		t.Fatal("exact-count formatting leaked")
	}
	if document, marshalErr := json.Marshal(statement); marshalErr == nil || document != nil {
		t.Fatal("exact-count JSON marshal did not fail closed")
	}
}
