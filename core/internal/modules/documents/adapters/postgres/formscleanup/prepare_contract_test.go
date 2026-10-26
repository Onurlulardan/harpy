package formscleanup

import (
	"strings"
	"testing"
)

func TestPrepareGatesLinkMutationOnCompletedAdoption(t *testing.T) {
	adoptionGate := strings.Index(prepareSQL, "JOIN adoption ON true")
	targetMutation := strings.Index(prepareSQL,
		"INSERT INTO documents_forms_cleanup_targets")
	linkMutation := strings.Index(prepareSQL,
		"UPDATE documents_links AS link")
	if adoptionGate < 0 || targetMutation < 0 || linkMutation < 0 ||
		adoptionGate > targetMutation || adoptionGate > linkMutation {
		t.Fatal("Forms cleanup can mutate links before completed adoption evidence")
	}
}
