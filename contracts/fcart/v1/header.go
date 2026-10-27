package fcartv1

import (
	"encoding/binary"

	"github.com/google/uuid"
)

type Header struct {
	Version           string
	ArtifactID        uuid.UUID
	AttemptGeneration uint64
	Format            Format
	ChunkSize         uint32
}

func NewHeader(
	artifactID uuid.UUID,
	attemptGeneration uint64,
	format Format,
) (Header, error) {
	header := Header{
		Version: Version, ArtifactID: artifactID,
		AttemptGeneration: attemptGeneration,
		Format:            format, ChunkSize: ChunkPlaintextBytes,
	}
	if err := header.Validate(); err != nil {
		return Header{}, err
	}
	return header, nil
}

func (header Header) Validate() error {
	if header.Version != Version || header.ArtifactID == uuid.Nil ||
		header.AttemptGeneration == 0 || header.ChunkSize != ChunkPlaintextBytes {
		return ErrInvalidHeader
	}
	if err := header.Format.Validate(); err != nil {
		return ErrInvalidHeader
	}
	return nil
}

func (header Header) encode() ([HeaderBytes]byte, error) {
	var encoded [HeaderBytes]byte
	if err := header.Validate(); err != nil {
		return encoded, err
	}
	formatCode, err := header.Format.wireCode()
	if err != nil {
		return encoded, ErrInvalidHeader
	}
	offset := copy(encoded[:], Version)
	offset += copy(encoded[offset:], header.ArtifactID[:])
	binary.BigEndian.PutUint64(encoded[offset:], header.AttemptGeneration)
	offset += 8
	encoded[offset] = formatCode
	offset++
	binary.BigEndian.PutUint32(encoded[offset:], header.ChunkSize)
	return encoded, nil
}

func decodeHeader(encoded [HeaderBytes]byte) (Header, error) {
	if string(encoded[:len(Version)]) != Version {
		return Header{}, ErrInvalidHeader
	}
	offset := len(Version)
	artifactID, err := uuid.FromBytes(encoded[offset : offset+16])
	if err != nil {
		return Header{}, ErrInvalidHeader
	}
	offset += 16
	attempt := binary.BigEndian.Uint64(encoded[offset:])
	offset += 8
	format, err := formatFromWire(encoded[offset])
	if err != nil {
		return Header{}, ErrInvalidHeader
	}
	offset++
	header := Header{
		Version: Version, ArtifactID: artifactID,
		AttemptGeneration: attempt, Format: format,
		ChunkSize: binary.BigEndian.Uint32(encoded[offset:]),
	}
	if err := header.Validate(); err != nil {
		return Header{}, err
	}
	return header, nil
}
