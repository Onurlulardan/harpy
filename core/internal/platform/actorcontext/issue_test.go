package actorcontext

import (
	"testing"

	"github.com/google/uuid"
)

func TestIssueRejectsIncompleteActorContext(t *testing.T) {
	if _, err := Issue(uuid.Nil, uuid.New()); err == nil {
		t.Fatal("nil actor was accepted")
	}
}
