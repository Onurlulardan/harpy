package compiler

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResealedCursorMetadataMismatchFailsClosed(t *testing.T) {
	base := fixtureBoundStatement(t, "forms", "display")
	tests := map[string]func(*RecordStatement){
		"direction": func(statement *RecordStatement) {
			statement.columns[1].sortDirection = api.DirectionDescending
		},
		"canonical-bound": func(statement *RecordStatement) {
			statement.columns[1].cursorCanonicalBytes++
		},
		"extraneous-sort": func(statement *RecordStatement) {
			statement.columns[0].sortOrdinal = 0
			statement.columns[0].sortDirection = api.DirectionAscending
			statement.columns[0].cursorCanonicalBytes = 256
		},
		"context-binding": func(statement *RecordStatement) {
			statement.cursorContext = fixtureBoundStatement(t, "crm", "display").cursorContext
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			statement := base
			statement.columns = append([]ResultColumn{}, base.columns...)
			mutate(&statement)
			seal, err := statementSeal(statement)
			if err != nil {
				t.Fatal(err)
			}
			statement.stateSeal = seal
			if !errors.Is(statement.Validate(), ErrContract) {
				t.Fatal("resealed cursor provenance mismatch was accepted")
			}
		})
	}
}
