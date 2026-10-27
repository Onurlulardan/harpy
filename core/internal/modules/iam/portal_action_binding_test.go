package iam

import (
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/google/uuid"
)

func TestPortalActionBindingRequiresExactCurrentEvidence(t *testing.T) {
	binding := portalActionFixtureBinding()
	decision, err := newPortalAccessDecision(binding)
	if err != nil || !samePortalActionBinding(binding, decision) {
		t.Fatalf("valid binding rejected: %v", err)
	}
	tests := map[string]func(*portaluser.AccessBinding){
		"form version": func(value *portaluser.AccessBinding) {
			value.FormVersionID = portalActionID(9)
		},
		"identifier": func(value *portaluser.AccessBinding) {
			value.IdentifierFieldID = portalActionID(9)
		},
		"roles": func(value *portaluser.AccessBinding) {
			value.RoleNames = []string{"attacker"}
		},
		"clock": func(value *portaluser.AccessBinding) {
			value.AuthorizationClock = 6
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			current := binding
			current.RoleNames = append([]string(nil), binding.RoleNames...)
			mutate(&current)
			if samePortalActionBinding(current, decision) {
				t.Fatalf("wrong %s accepted", name)
			}
		})
	}
}

func portalActionFixtureBinding() portaluser.AccessBinding {
	return portaluser.AccessBinding{
		PrincipalID: portalActionID(1), BindingID: portalActionID(2),
		TenantID: portalActionID(3), FormID: portalActionID(4),
		FormVersionID: portalActionID(5), SubjectRecordID: portalActionID(6),
		IdentifierFieldID: portalActionID(7), SubjectType: "employee",
		RoleNames: []string{"employee"}, TokenVersion: 5,
		AuthorizationVersion: 6, AuthorizationClock: 7, BindingRevision: 8,
	}
}

func portalActionID(seed byte) uuid.UUID {
	value := uuid.MustParse("018f47f4-7b1b-7000-8000-000000000000")
	value[15] = seed
	return value
}
