package iamactorv1

import (
	"strings"
	"testing"
)

func TestConsumeStatementIsStaticAndCallerTransactionScoped(t *testing.T) {
	required := []string{
		"iam_v1_actor_assertion_consume_v1",
		"pg_catalog.current_schema()::name",
		"$1", "$17",
	}
	for _, marker := range required {
		if !strings.Contains(consumeStatement, marker) {
			t.Fatalf("consume statement omits %s", marker)
		}
	}
	for _, forbidden := range []string{
		"BEGIN", "COMMIT", "ROLLBACK", "SET ROLE", "SET SESSION AUTHORIZATION",
		"fmt.", "%s",
	} {
		if strings.Contains(strings.ToUpper(consumeStatement), strings.ToUpper(forbidden)) {
			t.Fatalf("consume statement contains forbidden token %s", forbidden)
		}
	}
}
