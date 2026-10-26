package cutoverv1

import (
	"strings"
	"testing"
)

func TestAuthorityQuiescenceRejectsPolicyDrift(t *testing.T) {
	tests := map[string]func(*AuthorityAdoptionOperation){
		"principal": func(value *AuthorityAdoptionOperation) {
			value.QuiescencePolicy.Roles[0].Principal = "harpy_owner"
		},
		"temporary login": func(value *AuthorityAdoptionOperation) {
			value.QuiescencePolicy.Roles[0].TemporaryState.Login = true
		},
		"fixture connection limit": func(value *AuthorityAdoptionOperation) {
			value.QuiescencePolicy.Roles[0].PreState.ConnectionLimit = 1
		},
		"password state changed": func(value *AuthorityAdoptionOperation) {
			value.QuiescencePolicy.Roles[0].TemporaryState.PasswordState = "null"
		},
		"pre namespace": func(value *AuthorityAdoptionOperation) {
			value.QuiescencePolicy.ExpectedPreNamespaceSHA256 = strings.Repeat("f", 64)
		},
		"no termination": func(value *AuthorityAdoptionOperation) {
			value.QuiescencePolicy.TerminateSessions = false
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			operation := validAuthorityAdoptionOperation()
			operation.QuiescencePolicy.Roles = append(
				[]AuthorityQuiescedRolePolicy(nil), operation.QuiescencePolicy.Roles...,
			)
			mutate(&operation)
			if err := operation.Validate(); err == nil {
				t.Fatal("quiescence policy drift accepted")
			}
		})
	}
}

func TestAuthorityQuiescenceAcceptsNullPasswordAtRest(t *testing.T) {
	operation := validAuthorityAdoptionOperation()
	role := &operation.QuiescencePolicy.Roles[0]
	role.PreState.PasswordState = "null"
	role.TemporaryState.PasswordState = "null"
	role.PostState.PasswordState = "null"
	if err := operation.Validate(); err != nil {
		t.Fatalf("null-password authority quiescence rejected: %v", err)
	}
	receipt := validAuthorityAdoptionReceipt(operation)
	if err := receipt.Validate(); err != nil {
		t.Fatalf("null-password authority receipt rejected: %v", err)
	}
}

func TestAuthorityQuiescenceRejectsObservedDrift(t *testing.T) {
	operation := validAuthorityAdoptionOperation()
	tests := map[string]func(*AuthorityAdoptionReceipt){
		"remaining session": func(value *AuthorityAdoptionReceipt) {
			value.QuiescenceEvidence.Roles[0].RemainingSessionCount = 1
		},
		"password change": func(value *AuthorityAdoptionReceipt) {
			value.QuiescenceEvidence.Roles[0].PasswordVerifierUnchanged = false
		},
		"post namespace": func(value *AuthorityAdoptionReceipt) {
			value.QuiescenceEvidence.ObservedPostNamespaceSHA256 = strings.Repeat("e", 64)
		},
		"not restored": func(value *AuthorityAdoptionReceipt) {
			value.QuiescenceEvidence.RestoredBeforeCommit = false
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			receipt := validAuthorityAdoptionReceipt(operation)
			receipt.QuiescenceEvidence.Roles = append(
				[]AuthorityQuiescedRoleEvidence(nil), receipt.QuiescenceEvidence.Roles...,
			)
			mutate(&receipt)
			if err := receipt.Validate(); err == nil {
				t.Fatal("quiescence evidence drift accepted")
			}
		})
	}
}
