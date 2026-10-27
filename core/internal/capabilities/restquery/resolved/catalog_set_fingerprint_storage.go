package resolved

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type relationStorageFingerprintMember struct {
	relationID [16]byte
	storage    [sha256.Size]byte
}

func catalogStorageFingerprints(
	catalog api.FieldCatalog,
) ([sha256.Size]byte, [sha256.Size]byte, error) {
	storage, storageErr := catalog.Storage()
	resourceFingerprint, fingerprintErr := storage.FingerprintSHA256()
	relations, relationsErr := catalog.Relations()
	if storageErr != nil || fingerprintErr != nil || relationsErr != nil {
		return [sha256.Size]byte{}, [sha256.Size]byte{},
			fmt.Errorf("catalog storage mapping is invalid")
	}
	members := make([]relationStorageFingerprintMember, 0, len(relations))
	for _, relation := range relations {
		identifier, identifierErr := relation.RelationID()
		relationStorage, relationErr := relation.Storage()
		fingerprint, relationFingerprintErr := relationStorage.FingerprintSHA256()
		if identifierErr != nil || relationErr != nil || relationFingerprintErr != nil {
			return [sha256.Size]byte{}, [sha256.Size]byte{},
				fmt.Errorf("relation storage mapping is invalid")
		}
		member := relationStorageFingerprintMember{storage: fingerprint}
		copy(member.relationID[:], identifier[:])
		members = append(members, member)
	}
	slices.SortFunc(members, func(left, right relationStorageFingerprintMember) int {
		return bytes.Compare(left.relationID[:], right.relationID[:])
	})
	return resourceFingerprint, relationStorageFingerprint(members), nil
}

func relationStorageFingerprint(
	members []relationStorageFingerprintMember,
) [sha256.Size]byte {
	digest := sha256.New()
	_, _ = digest.Write([]byte("restquery_catalog_relation_storage_v1\x00"))
	var count [4]byte
	binary.BigEndian.PutUint32(count[:], uint32(len(members)))
	_, _ = digest.Write(count[:])
	for _, member := range members {
		_, _ = digest.Write(member.relationID[:])
		_, _ = digest.Write(member.storage[:])
	}
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}
