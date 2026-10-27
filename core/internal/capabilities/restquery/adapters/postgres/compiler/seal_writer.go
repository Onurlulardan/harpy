package compiler

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

type sealWriter struct {
	hash hash.Hash
	err  error
}

func newSealWriter(domain string) *sealWriter {
	writer := &sealWriter{hash: sha256.New()}
	writer.writeString(domain)
	return writer
}

func (writer *sealWriter) writeBytes(value []byte) {
	if writer == nil || writer.err != nil {
		return
	}
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	if _, err := writer.hash.Write(size[:]); err != nil {
		writer.err = err
		return
	}
	if _, err := writer.hash.Write(value); err != nil {
		writer.err = err
	}
}

func (writer *sealWriter) writeString(value string) { writer.writeBytes([]byte(value)) }

func (writer *sealWriter) writeInt(value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	writer.writeBytes(encoded[:])
}

func (writer *sealWriter) writeBool(value bool) {
	if value {
		writer.writeBytes([]byte{1})
		return
	}
	writer.writeBytes([]byte{0})
}

func (writer *sealWriter) sum() ([sha256.Size]byte, error) {
	if writer == nil || writer.err != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	var result [sha256.Size]byte
	copy(result[:], writer.hash.Sum(nil))
	if result == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, ErrContract
	}
	return result, nil
}
