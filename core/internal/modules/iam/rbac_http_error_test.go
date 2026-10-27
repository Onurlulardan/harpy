package iam

import (
	"errors"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
)

func TestRBACHTTPErrorPreservesPublicFailureSemantics(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"invalid", rbacdomain.ErrInvalidMutation, ErrRBACHTTPInvalid},
		{"denied", rbacdomain.ErrDenied, ErrRBACHTTPForbidden},
		{"not found", rbacdomain.ErrNotFound, ErrRBACHTTPNotFound},
		{"conflict", rbacdomain.ErrConflict, ErrRBACHTTPConflict},
		{"storage", rbacdomain.ErrStorage, ErrRBACHTTPUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := rbacHTTPError(test.err); !errors.Is(got, test.want) {
				t.Fatalf("rbacHTTPError() = %v, want %v", got, test.want)
			}
		})
	}
}
