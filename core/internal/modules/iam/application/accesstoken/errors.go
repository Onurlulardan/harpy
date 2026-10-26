package accesstoken

import "errors"

var (
	ErrInvalidProfile = errors.New("invalid access-token profile")
	ErrInvalidClaims  = errors.New("invalid access-token claims")
	ErrInvalidKey     = errors.New("invalid access-token key")
	ErrInvalidToken   = errors.New("invalid access token")
	ErrSignerClosed   = errors.New("access-token signer is closed")
)
