package application

import "testing"

func TestReadPreferencesMaterializesClosedDefaultMatrix(t *testing.T) {
	application, _, scope, _ := newIntentFixture(t)
	values, err := application.ReadPreferences(t.Context(), scope)
	if err != nil || len(values) != 18 {
		t.Fatalf("ReadPreferences() count=%d err=%v", len(values), err)
	}
	if values[0].Category == "" || values[0].Channel == "" {
		t.Fatal("default preference classification is empty")
	}
}
