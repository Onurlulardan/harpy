package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"github.com/google/uuid"
)

func (service *Service) ListInventory(
	ctx context.Context,
	recoverySetID uuid.UUID,
) ([]InventoryItem, [32]byte, error) {
	if !service.usable(ctx) || recoverySetID == uuid.Nil {
		return nil, [32]byte{}, ErrInvalid
	}
	items, err := service.store.ListInventory(ctx, recoverySetID)
	if err != nil {
		return nil, [32]byte{}, ErrUnavailable
	}
	return normalizeInventory(items)
}

func normalizeInventory(items []InventoryItem) ([]InventoryItem, [32]byte, error) {
	normalized := append([]InventoryItem(nil), items...)
	sort.Slice(normalized, func(left, right int) bool {
		return normalized[left].VersionID.String() <
			normalized[right].VersionID.String()
	})
	for _, item := range normalized {
		if !validInventoryItem(item) {
			return nil, [32]byte{}, ErrUnavailable
		}
	}
	return normalized, inventoryDigest(normalized), nil
}

func inventoryDigest(items []InventoryItem) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("documents-recovery-inventory.v1"))
	var size [8]byte
	for _, item := range items {
		for _, value := range [][]byte{
			[]byte(item.VersionID.String()), []byte(item.ObjectKey),
			[]byte(item.ObjectVersion), item.CiphertextSHA256[:],
			[]byte(item.KeyID),
		} {
			binary.BigEndian.PutUint64(size[:], uint64(len(value)))
			hash.Write(size[:])
			hash.Write(value)
		}
		binary.BigEndian.PutUint64(size[:], uint64(item.ObjectGeneration))
		hash.Write(size[:])
		binary.BigEndian.PutUint64(size[:], uint64(item.CiphertextBytes))
		hash.Write(size[:])
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validInventoryItem(item InventoryItem) bool {
	return item.VersionID != uuid.Nil && item.ObjectKey != "" &&
		item.ObjectVersion != "" && item.ObjectGeneration > 0 &&
		item.CiphertextSHA256 != [32]byte{} && item.CiphertextBytes > 0 &&
		item.KeyID != ""
}
