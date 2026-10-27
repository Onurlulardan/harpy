package download

import (
	"time"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func scanDescriptor(value row, tenantID domain.ID,
	notFound rowNotFound,
) (downloadapp.Descriptor, error) {
	var itemID, versionID uuid.UUID
	var name, mime, key, objectVersion, keyID string
	var generation, plainBytes, cipherBytes int64
	var plainHash, cipherHash, wrapped []byte
	var createdAt time.Time
	err := value.Scan(&itemID, &versionID, &name, &mime, &key, &objectVersion,
		&generation, &plainBytes, &cipherBytes, &plainHash, &cipherHash,
		&keyID, &wrapped, &createdAt)
	if err != nil {
		if notFound(err) {
			return downloadapp.Descriptor{}, downloadapp.ErrNotFound
		}
		return downloadapp.Descriptor{}, downloadapp.ErrUnavailable
	}
	if itemID == uuid.Nil || versionID == uuid.Nil || name == "" || mime == "" ||
		key == "" || objectVersion == "" || generation < 1 || plainBytes < 1 ||
		cipherBytes <= plainBytes || len(plainHash) != 32 || len(cipherHash) != 32 ||
		keyID == "" || len(wrapped) < 115 || createdAt.IsZero() {
		return downloadapp.Descriptor{}, downloadapp.ErrIntegrity
	}
	result := downloadapp.Descriptor{TenantID: tenantID, ItemID: domain.ID(itemID),
		VersionID: domain.ID(versionID), OriginalName: name, MIME: mime,
		Object: domain.ObjectReceipt{Key: key, Version: objectVersion, KeyID: keyID,
			Generation: uint64(generation), PlaintextBytes: plainBytes,
			CiphertextBytes: cipherBytes, WrappedDEK: append([]byte(nil), wrapped...)},
		CreatedAt: createdAt.UTC()}
	copy(result.Object.PlaintextSHA256[:], plainHash)
	copy(result.Object.CiphertextSHA256[:], cipherHash)
	return result, nil
}
