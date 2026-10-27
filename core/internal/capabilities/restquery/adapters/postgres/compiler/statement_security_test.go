package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestStatementSealsExecutionProvenance(t *testing.T) {
	statement := fixtureBoundStatement(t, "forms", "display")
	owner, ownerErr := statement.Owner()
	catalog, catalogErr := statement.Catalog()
	catalogOwner, catalogOwnerErr := catalog.Owner()
	tenant, tenantErr := statement.TenantID()
	actor, actorErr := statement.ActorID()
	identity, identityErr := statement.QueryIdentitySHA256()
	authorization, authorizationErr := statement.AuthorizationContextSHA256()
	mandatory, mandatoryErr := statement.MandatorySHA256()
	cursorContext, cursorErr := statement.CursorContext()
	contextPageSize, pageErr := cursorContext.PageSize()
	contextMandatory, contextMandatoryErr := cursorContext.MandatoryPredicateCanonical()
	if ownerErr != nil || catalogErr != nil || catalogOwnerErr != nil ||
		tenantErr != nil || actorErr != nil || identityErr != nil ||
		authorizationErr != nil || mandatoryErr != nil || cursorErr != nil ||
		pageErr != nil || contextMandatoryErr != nil || owner != "forms" ||
		catalogOwner != owner || tenant.String() != fixtureTenantID ||
		actor.String() != fixtureActorID || identity == [32]byte{} ||
		authorization == [32]byte{} || mandatory == [32]byte{} ||
		contextPageSize != 1 || len(contextMandatory) == 0 {
		t.Fatal("sealed execution provenance did not round trip")
	}

	other := fixtureBoundStatement(t, "crm", "display")
	if statement.stateSeal == other.stateSeal {
		t.Fatal("different catalog owners share a statement seal")
	}
	tampered := statement
	tampered.binding = other.binding
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("catalog owner swap was accepted")
	}
	tampered = statement
	tampered.binding.identity = other.binding.identity
	tampered.binding.identity[0] ^= 0xff
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("query identity tamper was accepted")
	}
	if !errors.Is(statement.ValidatePlan(apiPlanZero()), ErrContract) {
		t.Fatal("invalid plan was accepted by statement binding")
	}
}

func TestStatementAndColumnTamperingFailClosed(t *testing.T) {
	statement := fixtureBoundStatement(t, "forms", "display")
	otherContext := fixtureBoundStatement(t, "crm", "display").cursorContext
	for name, mutate := range map[string]func(*RecordStatement){
		"sql":       func(value *RecordStatement) { value.sql += " OR TRUE" },
		"arguments": func(value *RecordStatement) { value.arguments[0] = 999 },
		"output": func(value *RecordStatement) {
			value.columns[0].outputKey = "changed"
		},
		"type": func(value *RecordStatement) {
			value.columns[0].fieldType = fixtureDecimalType(t, 12, 4)
		},
		"class": func(value *RecordStatement) {
			value.class = planner.StatementClassExactCount
		},
		"context": func(value *RecordStatement) { value.cursorContext = otherContext },
		"direction": func(value *RecordStatement) {
			value.columns[1].sortDirection = api.DirectionDescending
		},
		"bound": func(value *RecordStatement) {
			value.columns[1].cursorCanonicalBytes++
		},
		"seal": func(value *RecordStatement) { value.stateSeal[0] ^= 0xff },
	} {
		t.Run(name, func(t *testing.T) {
			tampered := statement
			tampered.arguments = cloneArguments(statement.arguments)
			tampered.columns = append([]ResultColumn{}, statement.columns...)
			mutate(&tampered)
			if !errors.Is(tampered.Validate(), ErrContract) {
				t.Fatal("tampered statement was accepted")
			}
		})
	}
	column := statement.columns[0]
	column.outputKey = "_included"
	if !errors.Is(column.Validate(), ErrContract) {
		t.Fatal("reserved output key was accepted")
	}
	column = statement.columns[0]
	column.outputKey = strings.Repeat("x", maximumResultOutputKeyBytes+1)
	if !errors.Is(column.Validate(), ErrContract) {
		t.Fatal("oversized output key was accepted")
	}
	column = statement.columns[0]
	column.fieldID = [16]byte{1}
	if !errors.Is(column.Validate(), ErrContract) {
		t.Fatal("non-RFC field ID was accepted")
	}
}

func TestStatementDefensiveCopiesAndRedaction(t *testing.T) {
	base := fixtureBoundStatement(t, "forms", "display")
	values := []string{"OPEN", "CLOSED"}
	statement, err := newRecordStatement(
		"SELECT $1", []any{values}, base.columns, 1,
		planner.StatementClassRecord, base.countMode, base.cursorContext, base.binding,
	)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = "MUTATED"
	arguments, _ := statement.Arguments()
	arguments[0].([]string)[0] = "RETURNED_MUTATION"
	again, _ := statement.Arguments()
	if again[0].([]string)[0] != "OPEN" || statement.Validate() != nil {
		t.Fatal("statement argument ownership escaped")
	}
	if fmt.Sprint(statement) != compilerRedaction {
		t.Fatal("statement formatting leaked")
	}
	if document, marshalErr := json.Marshal(statement); marshalErr == nil || document != nil {
		t.Fatal("statement JSON marshal did not fail closed")
	}
}

func TestRecordStatementRejectsNonRecordClass(t *testing.T) {
	base := fixtureBoundStatement(t, "forms", "display")
	_, err := newRecordStatement(
		base.sql, base.arguments, base.columns, base.requestedSize,
		planner.StatementClassExactCount, base.countMode,
		base.cursorContext, base.binding,
	)
	if !errors.Is(err, ErrContract) {
		t.Fatal("exact-count class was accepted by record statement")
	}
}
