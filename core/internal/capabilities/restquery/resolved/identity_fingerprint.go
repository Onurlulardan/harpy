package resolved

import (
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func queryStateFingerprints(
	spec QuerySpec,
) ([sha256.Size]byte, [sha256.Size]byte, int, int, error) {
	identityEncoding, err := encodeQuerySpec(spec, false)
	if err != nil {
		return [sha256.Size]byte{}, [sha256.Size]byte{}, 0, 0, err
	}
	stateEncoding, err := encodeQuerySpec(spec, true)
	if err != nil {
		return [sha256.Size]byte{}, [sha256.Size]byte{}, 0, 0, err
	}
	return sha256.Sum256(identityEncoding), sha256.Sum256(stateEncoding),
		len(identityEncoding), len(stateEncoding), nil
}

func writeCatalogReference(writer *identityWriter, reference api.CatalogReference) {
	owner, ownerErr := reference.Owner()
	resourceID, resourceErr := reference.ResourceID()
	versionID, versionErr := reference.PublishedVersionID()
	hash, hashErr := reference.CatalogHash()
	hashBytes, bytesErr := hash.Bytes()
	if ownerErr != nil || resourceErr != nil || versionErr != nil ||
		hashErr != nil || bytesErr != nil {
		writer.err = fmt.Errorf("resolved REST Query catalog reference is invalid")
		return
	}
	writer.writeString(owner)
	writer.writeBytes(resourceID[:])
	writer.writeBytes(versionID[:])
	writer.writeBytes(hashBytes)
}
