package api

import "testing"

func TestValidateInvocationAcceptsNoArguments(t *testing.T) {
	if err := validateInvocation(nil); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"serve"}, {"--port", "8000"}} {
		if err := validateInvocation(arguments); err == nil {
			t.Fatalf("accepted arguments %v", arguments)
		}
	}
}
