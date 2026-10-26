package lifecyclecredential

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

type ID = [16]byte
type Purpose string

const (
	PurposeEmailVerification Purpose = "email_verification_v1"
	PurposePasswordReset     Purpose = "password_reset_v1"
	PurposePortalSetup       Purpose = "portal_setup_v1"
	PurposeInitialAdminSetup Purpose = "initial_admin_setup.v1"
	WireBytes                        = 66
)

var (
	ErrInvalid     = errors.New("lifecycle credential input is invalid")
	ErrUnavailable = errors.New("lifecycle credential runtime is unavailable")
)

type OneTimeKeyring interface {
	Active() (string, [32]byte, error)
	HMACKey(string) ([32]byte, error)
}

type TokenIssue struct {
	ID                        ID
	Selector                  [16]byte
	SecretHMAC, BindingSHA256 [32]byte
	HashKeyID                 string
}

func (issue TokenIssue) Validate() error {
	if issue.ID == [16]byte{} || issue.Selector == [16]byte{} ||
		issue.SecretHMAC == [32]byte{} || issue.BindingSHA256 == [32]byte{} ||
		issue.HashKeyID == "" {
		return ErrInvalid
	}
	return nil
}

func (purpose Purpose) Valid() bool {
	return purpose == PurposeEmailVerification || purpose == PurposePasswordReset ||
		purpose == PurposePortalSetup || purpose == PurposeInitialAdminSetup
}

func Digest(key [32]byte, tenantID ID, purpose Purpose,
	selector [16]byte, secret [32]byte,
) ([32]byte, error) {
	if key == [32]byte{} || tenantID == [16]byte{} || !purpose.Valid() ||
		selector == [16]byte{} || secret == [32]byte{} {
		return [32]byte{}, ErrInvalid
	}
	mac := hmac.New(sha256.New, key[:])
	for _, part := range [][]byte{[]byte("harpy:iam-lifecycle-secret:v1"),
		[]byte(purpose), tenantID[:], selector[:], secret[:]} {
		writePart(mac, part)
	}
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result, nil
}

func BindingSHA256(tenantID ID, purpose Purpose, selector [16]byte) ([32]byte, error) {
	if tenantID == [16]byte{} || !purpose.Valid() || selector == [16]byte{} {
		return [32]byte{}, ErrInvalid
	}
	hash := sha256.New()
	for _, part := range [][]byte{[]byte("harpy:iam-lifecycle-route:v1"),
		[]byte(purpose), tenantID[:], selector[:]} {
		writePart(hash, part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func writePart(writer interface{ Write([]byte) (int, error) }, value []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write(value)
}
