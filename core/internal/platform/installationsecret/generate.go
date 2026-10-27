package installationsecret

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

const (
	rawPasswordBytes  = 24
	wirePasswordBytes = 32
)

type TemporaryPassword struct {
	raw    [rawPasswordBytes]byte
	active bool
}

func GenerateTemporaryPassword() (*TemporaryPassword, error) {
	return generateTemporaryPassword(rand.Reader)
}

func generateTemporaryPassword(randomness io.Reader) (*TemporaryPassword, error) {
	password := &TemporaryPassword{active: true}
	if randomness == nil {
		password.Destroy()
		return nil, ErrInvalid
	}
	if _, err := io.ReadFull(randomness, password.raw[:]); err != nil ||
		password.raw == [rawPasswordBytes]byte{} {
		password.Destroy()
		return nil, ErrInterrupted
	}
	return password, nil
}

func (password *TemporaryPassword) Consume(consumer func([]byte) error) error {
	if password == nil || !password.active || consumer == nil {
		return ErrInvalid
	}
	wire := make([]byte, wirePasswordBytes)
	base64.RawURLEncoding.Encode(wire, password.raw[:])
	err := consumer(wire)
	clear(wire)
	password.Destroy()
	return err
}

func (password *TemporaryPassword) Destroy() {
	if password == nil {
		return
	}
	clear(password.raw[:])
	password.active = false
}
