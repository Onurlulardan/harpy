package work

import "testing"

func TestLegacyTaskUpdateIntentCanonicalizesV1RequestSemantics(t *testing.T) {
	taskID := workID(1)
	left, err := legacyTaskUpdateIntentHash(taskID, []byte(
		`{"title":null,"labels":[9007199254740993],"form_filter":{"ignored":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	right, err := legacyTaskUpdateIntentHash(taskID, []byte(
		`{ "labels" : [9007199254740993] }`))
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatal("null V1 pointer or ignored form_filter changed legacy intent")
	}
	labelsNull, err := legacyTaskUpdateIntentHash(taskID, []byte(`{"labels":null}`))
	if err != nil {
		t.Fatal(err)
	}
	empty, err := legacyTaskUpdateIntentHash(taskID, []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if labelsNull == empty {
		t.Fatal("labels null must remain an active V1 RawMessage intent")
	}
}
