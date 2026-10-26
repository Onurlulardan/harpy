package assignmentpostgres

import (
	"errors"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type assignmentStateError string

func (err assignmentStateError) Error() string    { return string(err) }
func (err assignmentStateError) SQLState() string { return string(err) }

func TestMapErrorPreservesAssignmentFailureSemantics(t *testing.T) {
	tests := []struct {
		state string
		want  error
	}{
		{"P0002", rbacdomain.ErrNotFound},
		{"40001", rbacdomain.ErrConflict},
		{"23505", rbacdomain.ErrConflict},
		{"42501", rbacdomain.ErrDenied},
		{"22023", rbacdomain.ErrInvalidQuery},
		{"XX000", rbacdomain.ErrStorage},
	}
	for _, test := range tests {
		got := mapError(assignmentStateError(test.state), rbacdomain.ErrInvalidQuery)
		if !errors.Is(got, test.want) {
			t.Fatalf("state %s mapped to %v, want %v", test.state, got, test.want)
		}
	}
}
