package systemtransfer

import "testing"

func TestNewWorkerRejectsTypedNilDependency(t *testing.T) {
	var scopes *fixtureScopes
	worker, err := NewWorker(scopes, &fixtureAuthorizer{},
		&fixtureIdentityLookup{}, &fixtureResourcePort{}, &fixtureInvitation{},
		&fixtureReceiptStore{})
	if err != ErrDependencies || worker != nil {
		t.Fatalf("worker=%v err=%v", worker, err)
	}
}
