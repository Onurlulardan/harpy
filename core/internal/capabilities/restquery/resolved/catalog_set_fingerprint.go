package resolved

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type catalogFingerprintMember struct {
	owner           string
	resource        [16]byte
	version         [16]byte
	hash            [sha256.Size]byte
	resourceStorage [sha256.Size]byte
	relationStorage [sha256.Size]byte
}

func catalogSetFingerprint(set api.CatalogSet) ([sha256.Size]byte, error) {
	catalogs, err := set.Catalogs()
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("resolved REST Query catalog set is invalid")
	}
	members := make([]catalogFingerprintMember, 0, len(catalogs))
	for _, catalog := range catalogs {
		member, memberErr := catalogFingerprintOf(catalog)
		if memberErr != nil {
			return [sha256.Size]byte{}, fmt.Errorf("resolved REST Query catalog set is invalid")
		}
		members = append(members, member)
	}
	slices.SortFunc(members, compareCatalogFingerprintMembers)
	hasher := sha256.New()
	var count [4]byte
	binary.BigEndian.PutUint32(count[:], uint32(len(members)))
	_, _ = hasher.Write(count[:])
	for _, member := range members {
		writeFingerprintMember(hasher, member)
	}
	var result [sha256.Size]byte
	copy(result[:], hasher.Sum(nil))
	return result, nil
}

func catalogFingerprintOf(
	catalog api.FieldCatalog,
) (catalogFingerprintMember, error) {
	reference, referenceErr := catalog.Reference()
	owner, ownerErr := reference.Owner()
	resource, resourceErr := reference.ResourceID()
	version, versionErr := reference.PublishedVersionID()
	hash, hashErr := reference.CatalogHash()
	hashBytes, bytesErr := hash.Bytes()
	resourceStorage, relationStorage, storageErr := catalogStorageFingerprints(catalog)
	if referenceErr != nil || ownerErr != nil || resourceErr != nil || versionErr != nil ||
		hashErr != nil || bytesErr != nil || storageErr != nil {
		return catalogFingerprintMember{}, fmt.Errorf("catalog reference is invalid")
	}
	member := catalogFingerprintMember{
		owner: owner, resourceStorage: resourceStorage,
		relationStorage: relationStorage,
	}
	copy(member.resource[:], resource[:])
	copy(member.version[:], version[:])
	copy(member.hash[:], hashBytes)
	return member, nil
}

func compareCatalogFingerprintMembers(left, right catalogFingerprintMember) int {
	if order := cmp.Compare(left.owner, right.owner); order != 0 {
		return order
	}
	if order := slices.Compare(left.resource[:], right.resource[:]); order != 0 {
		return order
	}
	if order := slices.Compare(left.version[:], right.version[:]); order != 0 {
		return order
	}
	return slices.Compare(left.hash[:], right.hash[:])
}

type byteWriter interface {
	Write([]byte) (int, error)
}

func writeFingerprintMember(writer byteWriter, member catalogFingerprintMember) {
	var ownerLength [4]byte
	binary.BigEndian.PutUint32(ownerLength[:], uint32(len(member.owner)))
	_, _ = writer.Write(ownerLength[:])
	_, _ = writer.Write([]byte(member.owner))
	_, _ = writer.Write(member.resource[:])
	_, _ = writer.Write(member.version[:])
	_, _ = writer.Write(member.hash[:])
	_, _ = writer.Write(member.resourceStorage[:])
	_, _ = writer.Write(member.relationStorage[:])
}
