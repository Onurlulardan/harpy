package fcartv1

import (
	"errors"
	"io"
)

func writePlaintextExact(destination io.Writer, plaintext []byte) error {
	written, err := destination.Write(plaintext)
	if err != nil {
		return errors.Join(ErrShortWrite, err)
	}
	if written != len(plaintext) {
		return ErrShortWrite
	}
	return nil
}
