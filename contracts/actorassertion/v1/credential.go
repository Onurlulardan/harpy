package actorassertionv1

import (
	"fmt"
	"io"
)

const redacted = "<redacted:iam-v1-actor-assertion>"

type Credential struct {
	wire string
}

func (credential Credential) Wire() (string, error) {
	if len(credential.wire) < 1 || len(credential.wire) > MaximumWireSize {
		return "", ErrContract
	}
	return credential.wire, nil
}

func (Credential) String() string               { return redacted }
func (Credential) GoString() string             { return redacted }
func (Credential) MarshalJSON() ([]byte, error) { return nil, ErrContract }
func (credential Credential) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, redacted)
}
