package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

const (
	initialSelectorBytes = 16
	initialSecretBytes   = 32
	initialCookieBytes   = 66
	initialCSRFWireBytes = 43
)

type InitialPasswordMaterial struct {
	selector [initialSelectorBytes]byte
	secret   [initialSecretBytes]byte
	csrf     [initialSecretBytes]byte
	active   bool
}

func generateInitialPasswordMaterial(randomness io.Reader) (*InitialPasswordMaterial, error) {
	if randomness == nil {
		return nil, ErrAuthenticationCrypto
	}
	result := &InitialPasswordMaterial{active: true}
	for _, target := range [][]byte{result.selector[:], result.secret[:], result.csrf[:]} {
		if _, err := io.ReadFull(randomness, target); err != nil {
			result.Destroy()
			return nil, ErrAuthenticationCrypto
		}
	}
	if result.selector == [initialSelectorBytes]byte{} ||
		result.secret == [initialSecretBytes]byte{} || result.csrf == [initialSecretBytes]byte{} {
		result.Destroy()
		return nil, ErrAuthenticationCrypto
	}
	return result, nil
}

func (material *InitialPasswordMaterial) RevealOnce(
	consumer func(cookie, csrf []byte) error,
) error {
	if material == nil || !material.active || consumer == nil {
		return ErrAuthenticationContract
	}
	cookie := make([]byte, initialCookieBytes)
	base64.RawURLEncoding.Encode(cookie[:22], material.selector[:])
	cookie[22] = '.'
	base64.RawURLEncoding.Encode(cookie[23:], material.secret[:])
	csrf := make([]byte, initialCSRFWireBytes)
	base64.RawURLEncoding.Encode(csrf, material.csrf[:])
	err := consumer(cookie, csrf)
	clear(cookie)
	clear(csrf)
	material.Destroy()
	return err
}

func (material *InitialPasswordMaterial) Destroy() {
	if material == nil {
		return
	}
	clear(material.selector[:])
	clear(material.secret[:])
	clear(material.csrf[:])
	material.active = false
}

func newInitialPasswordMaterial() (*InitialPasswordMaterial, error) {
	return generateInitialPasswordMaterial(rand.Reader)
}
