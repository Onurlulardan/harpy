package backupcrypto

import (
	"errors"
	"io"
)

const (
	Magic             = "FCBACKUP1"
	HeaderSchema      = "harpy-backup-header.v1"
	ChunkSize         = 1024 * 1024
	MaximumHeaderSize = 64 * 1024
)

var ErrInvalidArchive = errors.New("FCBACKUP1 archive is invalid")

type Context struct {
	EnvironmentID string
	RecoverySetID string
	ComponentID   string
	DRKeyID       string
	RecoveryKeyID string
}

type SealKeys struct {
	DRKEK          []byte
	RecoveryPublic []byte
	Random         io.Reader
}

type OpenKeys struct {
	DRKEK           []byte
	RecoveryPrivate []byte
}

type BoundContext struct {
	EnvironmentID string
	RecoverySetID string
	ComponentID   string
}

type OpenKeyResolver func(drKeyID, recoveryKeyID string) (OpenKeys, bool)

type Result struct {
	CiphertextBytes  int64
	CiphertextSHA256 string
	PlaintextBytes   int64
}
