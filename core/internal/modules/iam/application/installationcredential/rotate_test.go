package installationcredential

import (
	"context"
	"testing"
)

func TestRotateReturnsExactCredentialRotation(t *testing.T) {
	binding := rotationBindingFixture(t)
	store := &memoryStore{rotate: func(command RotateCommand) (RotateReceipt, error) {
		if command.Binding != binding || !command.Credential.Valid() {
			t.Fatal("rotate command drifted")
		}
		return RotateReceipt{
			OperationID: binding.OperationID, TenantID: binding.TenantID,
			UserID:               binding.FirstOwnerUserID,
			PreviousCredentialID: binding.CurrentCredentialID,
			CredentialID:         binding.NextCredentialID,
			CredentialVersion:    binding.NextCredentialVersion,
			CredentialRevision:   2, PreviousMarkerRevision: 2, MarkerRevision: 1,
		}, nil
	}}
	if _, err := Rotate(
		context.Background(), binding, store, credentialFixture(t),
	); err != nil {
		t.Fatal(err)
	}
}
