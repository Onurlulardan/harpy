package iam

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestSystemLayoutAccessTargetIsClosed(t *testing.T) {
	accepted := [][2]string{
		{"read", "system_resource_layout.get"},
		{"read", "system_resource_layout.list"},
		{"read", "system_resource_layout.resolve"},
		{"update", "system_resource_layout.create"},
		{"update", "system_resource_layout.update"},
		{"update", "system_resource_layout.delete"},
		{"update", "system_resource_layout.publish"},
	}
	for _, pair := range accepted {
		if _, valid := resolveSystemLayoutAccess(pair[0], pair[1]); !valid {
			t.Fatalf("valid target rejected: %q", pair)
		}
	}
	for _, pair := range [][2]string{
		{"delete", "system_resource_layout.delete"},
		{"update", "system_resource_layout.get"},
		{"read", "system_resource_layout.publish"},
		{"update", "system_resource_layout.unknown"},
	} {
		if _, valid := resolveSystemLayoutAccess(pair[0], pair[1]); valid {
			t.Fatalf("unknown target accepted: %q", pair)
		}
	}
}

func TestSystemLayoutDecisionOwnsExactAdmission(t *testing.T) {
	tenantID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000001")
	actorID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000002")
	admission, err := systemLayoutAdmission(tenantID, actorID, 3, 4, 7)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := NewSystemLayoutAccessDecision(admission)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := decision.Admission()
	if err != nil || stored.Validate() != nil {
		t.Fatalf("stored admission = %#v: %v", stored, err)
	}
	if _, err := NewSystemLayoutAccessDecision(publicresource.Admission{}); err == nil {
		t.Fatal("zero admission accepted")
	}
}

func TestSystemLayoutAccessRejectsInvalidRequestBeforeTokenWork(t *testing.T) {
	_, err := (*SystemLayoutAccessRuntime)(nil).Verify(
		context.Background(), SystemLayoutAccessRequest{},
	)
	if err != ErrSystemLayoutAccessUnavailable {
		t.Fatalf("nil runtime error = %v", err)
	}
	runtime := &SystemLayoutAccessRuntime{}
	_, err = runtime.Verify(context.Background(), SystemLayoutAccessRequest{})
	if err != ErrSystemLayoutAccessUnavailable {
		t.Fatalf("invalid resource error = %v", err)
	}
}
