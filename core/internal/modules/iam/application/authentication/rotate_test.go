package authentication

import (
	"context"
	"testing"
)

type trackedPresentedRefresh struct{ destroyed bool }

func (*trackedPresentedRefresh) Selector() ([16]byte, error) { return [16]byte{}, nil }
func (*trackedPresentedRefresh) VerifyHMAC([]byte, [32]byte) (bool, error) {
	return false, nil
}
func (token *trackedPresentedRefresh) Destroy() { token.destroyed = true }

func TestRotateFailsClosedAndDestroysCredentialOutsideScope(t *testing.T) {
	presented := &trackedPresentedRefresh{}
	rotation, err := (Authenticator{}).rotateWithClient(
		context.Background(), CredentialScope{}, RotateRequest{Presented: presented},
	)
	if err == nil || rotation.valid || !presented.destroyed {
		t.Fatalf("rotation valid = %t, destroyed = %t, err = %v",
			rotation.valid, presented.destroyed, err)
	}
}

func TestRotateBrowserFailsClosedAndDestroysCredentialOutsideScope(t *testing.T) {
	presented := &trackedPresentedRefresh{}
	rotation, err := (Authenticator{}).RotateBrowser(
		context.Background(), CredentialScope{}, presented,
	)
	if err == nil || rotation.valid || !presented.destroyed {
		t.Fatalf("rotation valid = %t, destroyed = %t, err = %v",
			rotation.valid, presented.destroyed, err)
	}
}
