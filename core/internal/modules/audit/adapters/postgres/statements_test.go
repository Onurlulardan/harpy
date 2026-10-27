package postgres

import (
	"strings"
	"testing"
)

func TestOwnedStatementsAreEmbeddedFromStaticSQL(t *testing.T) {
	for name, statement := range map[string]string{
		"append": appendSQL, "list": listSQL, "get": getSQL, "stats": statsSQL,
	} {
		if strings.TrimSpace(statement) == "" || !strings.HasPrefix(statement, "SELECT") {
			t.Fatalf("%s statement was not embedded", name)
		}
	}
}
