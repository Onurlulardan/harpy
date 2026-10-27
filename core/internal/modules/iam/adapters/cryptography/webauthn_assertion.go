package cryptography

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"

type WebAuthnAssertion struct{ secureMaterial }

func NewWebAuthnAssertion(value []byte) (*WebAuthnAssertion, error) {
	material, err := newSecureMaterial(value, 2, 64*1024)
	if err != nil {
		return nil, authentication.ErrAuthenticationContract
	}
	return &WebAuthnAssertion{secureMaterial: material}, nil
}

func (assertion *WebAuthnAssertion) Bytes() ([]byte, error) {
	value, err := assertion.copy()
	if err != nil {
		return nil, authentication.ErrAuthenticationContract
	}
	return value, nil
}
