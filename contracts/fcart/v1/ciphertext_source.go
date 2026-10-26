package fcartv1

import (
	"crypto/sha256"
	"errors"
	"hash"
	"io"
)

type ciphertextSource struct {
	source    io.Reader
	hasher    hash.Hash
	remaining uint64
}

func newCiphertextSource(
	source io.Reader,
	expectedBytes uint64,
) *ciphertextSource {
	return &ciphertextSource{
		source: source, hasher: sha256.New(), remaining: expectedBytes,
	}
}

func (source *ciphertextSource) readExact(destination []byte) error {
	if uint64(len(destination)) > source.remaining {
		return ErrTruncated
	}
	read, err := io.ReadFull(source.source, destination)
	if read > 0 {
		_, _ = source.hasher.Write(destination[:read])
		source.remaining -= uint64(read)
	}
	if err != nil {
		return errors.Join(ErrTruncated, err)
	}
	return nil
}

func (source *ciphertextSource) requireEOF() error {
	if source.remaining != 0 {
		return ErrTruncated
	}
	var extra [1]byte
	read, err := io.ReadFull(source.source, extra[:])
	if read != 0 || err == nil {
		return ErrTrailingData
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	return errors.Join(ErrCiphertextRead, err)
}

func (source *ciphertextSource) digest() [32]byte {
	var digest [32]byte
	copy(digest[:], source.hasher.Sum(nil))
	return digest
}
