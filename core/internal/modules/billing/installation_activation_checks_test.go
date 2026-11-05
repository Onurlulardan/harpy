package billing

import (
	"errors"
	"testing"
)

func TestCheckBillingInstallationActivation(t *testing.T) {
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
			err := checkBillingInstallationActivation(
				"assert billing installation activation", test.valid, test.err,
			)
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

func TestCheckBillingInstallationUsage(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	if err := checkBillingInstallationUsage(1, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := checkBillingInstallationUsage(0, false, nil); !errors.Is(
		err, ErrInstallationUsageNotOpened,
	) {
		t.Fatalf("err=%v", err)
	}
	if err := checkBillingInstallationUsage(0, false, databaseErr); !errors.Is(
		err, databaseErr,
	) {
		t.Fatalf("err=%v", err)
	}
}
