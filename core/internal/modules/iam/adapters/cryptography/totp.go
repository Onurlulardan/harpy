package cryptography

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

type TOTPSecretDecryptor interface {
	DecryptTOTP(
		context.Context, string, []byte, []byte, []byte, [32]byte,
	) ([32]byte, error)
}

type TOTP struct{ decryptor TOTPSecretDecryptor }

func NewTOTP(decryptor TOTPSecretDecryptor) (TOTP, error) {
	if nilDependency(decryptor) {
		return TOTP{}, authentication.ErrAuthenticationContract
	}
	return TOTP{decryptor: decryptor}, nil
}

func (provider TOTP) Verify(
	ctx context.Context,
	code authentication.TOTPCode,
	envelope authentication.TOTPEnvelope,
	now time.Time,
) (int64, error) {
	concrete, ok := code.(*TOTPCode)
	if nilDependency(provider.decryptor) || ctx == nil || ctx.Err() != nil ||
		now.IsZero() || !ok || concrete == nil {
		return 0, authentication.ErrAuthenticationContract
	}
	digits, err := concrete.Digits()
	if err != nil {
		return 0, authentication.ErrAuthenticationContract
	}
	defer clear(digits[:])
	secret, err := provider.decryptor.DecryptTOTP(
		ctx, envelope.KMSKeyID, envelope.Nonce, envelope.Ciphertext,
		envelope.AuthenticationTag, envelope.AADSHA256,
	)
	if err != nil {
		return 0, authentication.ErrAuthenticationCrypto
	}
	defer clear(secret[:])
	current := now.UTC().Unix() / 30
	for _, step := range []int64{current, current - 1, current + 1} {
		expected := totpDigits(secret, step)
		matched := subtle.ConstantTimeCompare(digits[:], expected[:]) == 1
		clear(expected[:])
		if matched {
			return step, nil
		}
	}
	return 0, authentication.ErrInvalidCredentials
}

func totpDigits(secret [32]byte, step int64) [6]byte {
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(step))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write(counter[:])
	digest := mac.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff
	value %= 1_000_000
	var result [6]byte
	for index := 5; index >= 0; index-- {
		result[index] = byte(value%10) + '0'
		value /= 10
	}
	clear(digest)
	return result
}
