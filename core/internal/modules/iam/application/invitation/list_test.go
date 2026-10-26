package invitation

import "testing"

func TestListQueryRejectsDuplicateStateAndUnboundedPage(t *testing.T) {
	if err := (ListQuery{States: []State{StatePending}, Limit: 100}).Validate(); err != nil {
		t.Fatalf("valid list query: %v", err)
	}
	if err := (ListQuery{
		States: []State{StatePending, StatePending}, Limit: 10,
	}).Validate(); err == nil {
		t.Fatal("duplicate invitation state was accepted")
	}
	if err := (ListQuery{Limit: 101}).Validate(); err == nil {
		t.Fatal("unbounded invitation page was accepted")
	}
}
