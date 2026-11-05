package systemresourcelayout

import (
	"errors"
	"testing"
)

func TestCheckInstallationBaselineActivation(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	for _, test := range []struct {
		name  string
		valid bool
		err   error
		want  error
	}{
		{name: "verified", valid: true},
		{name: "denied", want: ErrInstallationActivationUnverified},
		{name: "database error", err: databaseErr, want: databaseErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := checkInstallationBaselineActivation(test.valid, test.err)
			if test.want == nil && err != nil {
				t.Fatal(err)
			}
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("err=%v want wrapped %v", err, test.want)
			}
			if err != nil && err.Error() == "%!w(<nil>)" {
				t.Fatal("nil error was formatted as a wrapped error")
			}
		})
	}
}
