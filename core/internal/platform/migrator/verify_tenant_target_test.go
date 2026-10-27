package migrator

import (
	"errors"
	"testing"
)

type tenantVerificationCloseFixture struct {
	error error
	calls int
}

func (fixture *tenantVerificationCloseFixture) Close() error {
	fixture.calls++
	return fixture.error
}

func TestVerifyTenantTargetJoinsDatabaseCloseError(t *testing.T) {
	verificationErr := errors.New("verification failed")
	closeErr := errors.New("close failed")
	resultErr := error(verificationErr)
	fixture := &tenantVerificationCloseFixture{error: closeErr}

	joinTenantVerificationDatabaseClose(&resultErr, fixture)

	if fixture.calls != 1 {
		t.Fatalf("close calls = %d, want 1", fixture.calls)
	}
	if !errors.Is(resultErr, verificationErr) || !errors.Is(resultErr, closeErr) {
		t.Fatalf("joined error = %v", resultErr)
	}
}

func TestVerifyTenantTargetReturnsDatabaseCloseErrorAfterSuccess(t *testing.T) {
	closeErr := errors.New("close failed")
	var resultErr error
	fixture := &tenantVerificationCloseFixture{error: closeErr}

	joinTenantVerificationDatabaseClose(&resultErr, fixture)

	if !errors.Is(resultErr, closeErr) {
		t.Fatalf("close error = %v", resultErr)
	}
}
