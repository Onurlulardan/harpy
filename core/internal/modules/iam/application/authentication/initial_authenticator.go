package authentication

import (
	"crypto/rand"
	"io"
)

type InitialPasswordAuthenticator struct {
	base         Authenticator
	port         InitialPasswordPort
	continuation InitialPasswordKeyring
	csrf         InitialPasswordKeyring
	randomness   io.Reader
	valid        bool
}

func NewInitialPasswordAuthenticator(
	base Authenticator,
	port InitialPasswordPort,
	continuation InitialPasswordKeyring,
	csrf InitialPasswordKeyring,
) (InitialPasswordAuthenticator, error) {
	if !base.valid || nilInterface(port) || nilInterface(continuation) || nilInterface(csrf) {
		return InitialPasswordAuthenticator{}, ErrAuthenticationContract
	}
	return InitialPasswordAuthenticator{base: base, port: port,
		continuation: continuation, csrf: csrf, randomness: rand.Reader, valid: true}, nil
}
