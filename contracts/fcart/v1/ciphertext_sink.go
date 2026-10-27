package fcartv1

import (
	"crypto/sha256"
	"errors"
	"hash"
	"io"
)

type ciphertextSink struct {
	destination io.Writer
	hasher      hash.Hash
	written     uint64
	limit       uint64
}

func newCiphertextSink(destination io.Writer, limit uint64) *ciphertextSink {
	return &ciphertextSink{
		destination: destination,
		hasher:      sha256.New(),
		limit:       limit,
	}
}

func (sink *ciphertextSink) writeExact(value []byte) error {
	if uint64(len(value)) > sink.limit-sink.written {
		return ErrCiphertextLimit
	}
	written, err := sink.destination.Write(value)
	if written < 0 || written > len(value) {
		return ErrShortWrite
	}
	if written > 0 {
		_, _ = sink.hasher.Write(value[:written])
		sink.written += uint64(written)
	}
	if err != nil {
		return errors.Join(ErrShortWrite, err)
	}
	if written != len(value) {
		return ErrShortWrite
	}
	return nil
}

func (sink *ciphertextSink) digest() [32]byte {
	var digest [32]byte
	copy(digest[:], sink.hasher.Sum(nil))
	return digest
}
