package domain

import (
	"errors"
	"time"
)

var ErrVersionInvalid = errors.New("documents version input is invalid")

type VersionState string

const (
	VersionReady         VersionState = "ready"
	VersionDeletePending VersionState = "delete_pending"
	VersionPurged        VersionState = "purged"
)

type ObjectReceipt struct {
	Key, Version, KeyID               string
	Generation                        uint64
	PlaintextBytes, CiphertextBytes   int64
	PlaintextSHA256, CiphertextSHA256 [32]byte
	WrappedDEK                        []byte
}

type Version struct {
	ID, TenantID, ItemID ID
	Number               uint64
	State                VersionState
	Declaration          FileDeclaration
	DetectedMIME         string
	Object               ObjectReceipt
	ScanRevision         string
	ScannedAt, CreatedAt time.Time
}

func NewReadyVersion(id, tenantID, itemID ID, number uint64, declaration FileDeclaration,
	detectedMIME string, object ObjectReceipt, scanRevision string, now time.Time,
) (Version, error) {
	expectedKey, keyErr := ContentObjectKey(tenantID, itemID, id)
	if !id.UUIDv7() || !tenantID.Valid() || !itemID.UUIDv7() || number == 0 ||
		ValidateFile(declaration) != nil || detectedMIME == "" || object.Key != expectedKey ||
		keyErr != nil || !validObjectReceipt(object, declaration.Size) || scanRevision == "" ||
		len(scanRevision) > 256 || !validInstant(now) {
		return Version{}, ErrVersionInvalid
	}
	now = now.UTC()
	return Version{ID: id, TenantID: tenantID, ItemID: itemID, Number: number,
		State: VersionReady, Declaration: declaration, DetectedMIME: detectedMIME,
		Object: object, ScanRevision: scanRevision, ScannedAt: now, CreatedAt: now}, nil
}

func (version Version) Downloadable() bool {
	return version.State == VersionReady && version.ID.UUIDv7() &&
		version.Object.PlaintextSHA256 != [32]byte{}
}

func validObjectReceipt(value ObjectReceipt, expectedBytes int64) bool {
	return value.Version != "" && len(value.Version) <= 1024 && value.KeyID != "" &&
		len(value.KeyID) <= 64 && value.Generation > 0 && value.PlaintextBytes == expectedBytes &&
		value.CiphertextBytes > value.PlaintextBytes && value.PlaintextSHA256 != [32]byte{} &&
		value.CiphertextSHA256 != [32]byte{} && len(value.WrappedDEK) >= 115 &&
		len(value.WrappedDEK) <= 256
}
