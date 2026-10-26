package customerprovisioning

import (
	"errors"
	"testing"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/jackc/pgx/v5"
)

type stateError string

func (err stateError) Error() string    { return string(err) }
func (err stateError) SQLState() string { return string(err) }

func TestClassifyStableErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  error
		want error
	}{
		{name: "conflict", got: stateError("23505"), want: provisioning.ErrConflict},
		{name: "forbidden", got: stateError("42501"), want: provisioning.ErrForbidden},
		{name: "no rows", got: pgx.ErrNoRows, want: provisioning.ErrNotFound},
		{name: "unavailable", got: errors.New("database failed"), want: provisioning.ErrUnavailable},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := classify(test.got, func(err error) bool {
				return errors.Is(err, pgx.ErrNoRows)
			}); !errors.Is(got, test.want) {
				t.Fatalf("classify() = %v, want %v", got, test.want)
			}
		})
	}
}
