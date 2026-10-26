package sqlidentifier_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestTokenHasNoPublicRenderMethod(t *testing.T) {
	t.Parallel()
	if _, found := reflect.TypeFor[sqlidentifier.Token]().MethodByName("AppendSQL"); found {
		t.Fatal("Token exposes AppendSQL and can bypass the compiler call-site guard")
	}
}

func TestAppendSQLWritesOneQuotedIdentifier(t *testing.T) {
	t.Parallel()
	token, err := sqlidentifier.MintStatic("forms_definitions")
	if err != nil {
		t.Fatal(err)
	}
	var builder strings.Builder
	builder.WriteString("SELECT * FROM ")
	if err := sqlidentifier.AppendSQL(token, &builder); err != nil {
		t.Fatal(err)
	}
	if got, want := builder.String(), `SELECT * FROM "forms_definitions"`; got != want {
		t.Fatalf("SQL = %q, want %q", got, want)
	}
}

func TestAppendSQLRejectsZeroTokenAndNilBuilder(t *testing.T) {
	t.Parallel()
	var zero sqlidentifier.Token
	var builder strings.Builder
	builder.WriteString("stable")
	if err := sqlidentifier.AppendSQL(zero, &builder); err == nil || builder.String() != "stable" {
		t.Fatalf("zero token changed builder or succeeded: %q, %v", builder.String(), err)
	}
	token, err := sqlidentifier.MintStatic("forms_definitions")
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlidentifier.AppendSQL(token, nil); err == nil {
		t.Fatal("nil builder was accepted")
	}
}
