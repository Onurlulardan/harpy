package platformauthority

import (
	"errors"
	"strings"
	"testing"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func TestDevelopmentRecoveryTrustErrorNamesTheCause(t *testing.T) {
	path := "/run/dev-authority/current/recovery-trust-bundle.v1.json"
	parseErr := developmentRecoveryTrustError(path, "environment", "digest",
		recoveryv1.Bundle{}, errors.New("recovery signing bundle is invalid"))
	if parseErr == nil || !strings.Contains(parseErr.Error(), path) ||
		!strings.Contains(parseErr.Error(), "recovery signing bundle is invalid") {
		t.Fatal(parseErr)
	}
	environmentErr := developmentRecoveryTrustError(path, "expected-environment",
		"stored", recoveryv1.Bundle{}, nil)
	if environmentErr == nil || !strings.Contains(environmentErr.Error(), "expected-environment") {
		t.Fatal(environmentErr)
	}
	digestErr := developmentRecoveryTrustError(path, "", "stored-digest",
		recoveryv1.Bundle{}, nil)
	if digestErr == nil || !strings.Contains(digestErr.Error(), "stored-digest") ||
		!strings.Contains(digestErr.Error(), path) {
		t.Fatal(digestErr)
	}
	if err := developmentRecoveryTrustError(path, "", "", recoveryv1.Bundle{}, nil); err != nil {
		t.Fatal(err)
	}
}
