package work

import "testing"

func TestCompatibilityCalendarListDoesNotRequireForms(t *testing.T) {
	if compatibilityReadNeedsForms("calendar") {
		t.Fatal("v1 calendar list is user-scoped events, not a forms snapshot")
	}
	if !compatibilityReadNeedsForms("boards") || !compatibilityReadNeedsForms("tasks") {
		t.Fatal("board and task compatibility reads still require forms")
	}
}
