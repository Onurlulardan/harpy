package postgres

import (
	"strings"
	"testing"
)

func TestAuthorizeUsesEffectiveBoardRoleWithoutExternalOwnerSQL(t *testing.T) {
	for _, fragment := range []string{
		"supplied_principals", "membership.board_role", "FOR SHARE OF membership",
		"work.task.%", "work.board.update",
	} {
		if !strings.Contains(authorizeMutationSQL, fragment) {
			t.Fatalf("authorization query missing %q", fragment)
		}
	}
	lower := strings.ToLower(authorizeMutationSQL)
	for _, forbidden := range []string{"from iam_", "join iam_"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("authorization directly reads IAM via %q", forbidden)
		}
	}
}
