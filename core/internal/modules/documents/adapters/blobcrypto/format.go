package blobcrypto

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const (
	FormatID  = "FCDOC1"
	ChunkSize = 8 * 1024 * 1024
	headerLen = 6 + 16 + 16 + 16 + 8 + 4 + 8
)

var (
	ErrInvalidFormat    = errors.New("documents encrypted blob format is invalid")
	ErrIdentityMismatch = errors.New("documents encrypted blob identity mismatches")
	ErrContentTampered  = errors.New("documents encrypted blob authentication failed")
)

type Metadata struct {
	TenantID       domain.ID
	DocumentID     domain.ID
	VersionID      domain.ID
	Generation     uint64
	PlaintextBytes int64
}

type Receipt struct {
	Metadata         Metadata
	PlaintextSHA256  [32]byte
	CiphertextSHA256 [32]byte
}

func (metadata Metadata) valid() bool {
	return metadata.TenantID.Valid() && metadata.DocumentID.UUIDv7() &&
		metadata.VersionID.UUIDv7() && metadata.Generation > 0 &&
		metadata.PlaintextBytes >= 0 &&
		metadata.PlaintextBytes <= domain.MaximumArchiveArtifactPlaintext
}

func encodeHeader(metadata Metadata) ([]byte, error) {
	if !metadata.valid() {
		return nil, ErrInvalidFormat
	}
	header := make([]byte, headerLen)
	copy(header, FormatID)
	copy(header[6:22], metadata.TenantID[:])
	copy(header[22:38], metadata.DocumentID[:])
	copy(header[38:54], metadata.VersionID[:])
	binary.BigEndian.PutUint64(header[54:62], metadata.Generation)
	binary.BigEndian.PutUint32(header[62:66], ChunkSize)
	binary.BigEndian.PutUint64(header[66:74], uint64(metadata.PlaintextBytes))
	return header, nil
}

func readHeader(reader io.Reader) (Metadata, []byte, error) {
	header := make([]byte, headerLen)
	if _, err := io.ReadFull(reader, header); err != nil || string(header[:6]) != FormatID ||
		binary.BigEndian.Uint32(header[62:66]) != ChunkSize {
		return Metadata{}, nil, ErrInvalidFormat
	}
	metadata := Metadata{Generation: binary.BigEndian.Uint64(header[54:62]),
		PlaintextBytes: int64(binary.BigEndian.Uint64(header[66:74]))}
	copy(metadata.TenantID[:], header[6:22])
	copy(metadata.DocumentID[:], header[22:38])
	copy(metadata.VersionID[:], header[38:54])
	if !metadata.valid() {
		return Metadata{}, nil, ErrInvalidFormat
	}
	return metadata, header, nil
}

func chunkAAD(header []byte, index uint32, size uint32) []byte {
	result := make([]byte, len(header)+8)
	copy(result, header)
	binary.BigEndian.PutUint32(result[len(header):], index)
	binary.BigEndian.PutUint32(result[len(header)+4:], size)
	return result
}

func sameIdentity(left, right Metadata) bool {
	return left.TenantID == right.TenantID && left.DocumentID == right.DocumentID &&
		left.VersionID == right.VersionID && left.Generation == right.Generation &&
		left.PlaintextBytes == right.PlaintextBytes
}
