package iam

import (
	"errors"
	"testing"
)

func TestCheckInstallationActivationAssertion(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	tests := []struct {
		name  string
		valid bool
		err   error
		want  error
	}{
		{name: "verified", valid: true},
		{name: "denied", want: ErrInstallationActivationUnverified},
		{name: "database error", err: databaseErr, want: databaseErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkInstallationActivationAssertion(test.valid, test.err)
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

func TestCheckInstallationSecurityEpoch(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	if err := checkInstallationSecurityEpoch(true, nil); err != nil {
		t.Fatal(err)
	}
	if err := checkInstallationSecurityEpoch(false, nil); !errors.Is(
		err, ErrInstallationSecurityEpochNotReady,
	) {
		t.Fatalf("err=%v", err)
	}
	if err := checkInstallationSecurityEpoch(false, databaseErr); !errors.Is(
		err, databaseErr,
	) {
		t.Fatalf("err=%v", err)
	}
}

func TestCheckInstallationDirectoryRevision(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	if err := checkInstallationDirectoryRevision(1, nil); err != nil {
		t.Fatal(err)
	}
	if err := checkInstallationDirectoryRevision(0, nil); !errors.Is(
		err, ErrInstallationDirectoryNotInitialized,
	) {
		t.Fatalf("err=%v", err)
	}
	if err := checkInstallationDirectoryRevision(0, databaseErr); !errors.Is(
		err, databaseErr,
	) {
		t.Fatalf("err=%v", err)
	}
}
