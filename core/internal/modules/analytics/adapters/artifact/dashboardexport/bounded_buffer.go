package dashboardexportartifact

import (
	"crypto/sha256"
	"errors"
	"hash"
	"io"
)

var errPlaintextLimit = errors.New("analytics dashboard export plaintext limit exceeded")

type plaintextEvidenceWriter struct {
	destination io.Writer
	digest      hash.Hash
	maximum     int64
	written     int64
}

func newPlaintextEvidenceWriter(
	destination io.Writer, maximum int64,
) *plaintextEvidenceWriter {
	return &plaintextEvidenceWriter{destination: destination,
		digest: sha256.New(), maximum: maximum}
}

func (writer *plaintextEvidenceWriter) Write(value []byte) (int, error) {
	if writer == nil || writer.destination == nil || writer.digest == nil ||
		int64(len(value)) > writer.maximum-writer.written {
		return 0, errPlaintextLimit
	}
	written, err := writer.destination.Write(value)
	if written > 0 {
		_, _ = writer.digest.Write(value[:written])
		writer.written += int64(written)
	}
	return written, err
}

func (writer *plaintextEvidenceWriter) evidence() (int64, [sha256.Size]byte) {
	var digest [sha256.Size]byte
	if writer == nil || writer.digest == nil {
		return 0, digest
	}
	copy(digest[:], writer.digest.Sum(nil))
	return writer.written, digest
}
