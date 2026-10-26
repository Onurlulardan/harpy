package authentication

import (
	"context"
	"testing"
)

type trackedRecoveryCode struct{ destroyed bool }

func (*trackedRecoveryCode) KeyID() (string, error)        { return "recovery-v1", nil }
func (*trackedRecoveryCode) HMAC([]byte) ([32]byte, error) { return [32]byte{}, nil }
func (code *trackedRecoveryCode) Destroy()                 { code.destroyed = true }

func TestConsumeRecoveryFailsClosedAndDestroysCodeOutsideScope(t *testing.T) {
	code := &trackedRecoveryCode{}
	_, err := (Authenticator{}).ConsumeRecovery(
		context.Background(), CredentialScope{}, RecoveryRequest{Code: code},
	)
	if err == nil || !code.destroyed {
		t.Fatalf("destroyed = %t, err = %v", code.destroyed, err)
	}
}
