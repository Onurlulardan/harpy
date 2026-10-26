package worker

import "testing"

func TestValidateInvocationAcceptsNoArguments(t *testing.T) {
	if err := validateInvocation(nil); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"run"}, {"--once"}} {
		if err := validateInvocation(arguments); err == nil {
			t.Fatalf("accepted arguments %v", arguments)
		}
	}
}
