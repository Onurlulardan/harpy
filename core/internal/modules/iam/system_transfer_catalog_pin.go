package iam

import (
	"crypto/subtle"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	querycatalog "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/google/uuid"
)

type systemTransferCatalogPin struct {
	versionID uuid.UUID
	sha256    [32]byte
}

func systemTransferPublishedPin(
	published querycatalog.Published,
) (systemTransferCatalogPin, error) {
	reference, err := published.Reference()
	if err != nil {
		return systemTransferCatalogPin{}, ErrSystemTransferUnavailable
	}
	version, err := reference.PublishedVersionID()
	if err != nil || version == uuid.Nil {
		return systemTransferCatalogPin{}, ErrSystemTransferUnavailable
	}
	hash, err := reference.CatalogHash()
	if err != nil {
		return systemTransferCatalogPin{}, ErrSystemTransferUnavailable
	}
	bytes, err := hash.Bytes()
	if err != nil || len(bytes) != 32 {
		return systemTransferCatalogPin{}, ErrSystemTransferUnavailable
	}
	var digest [32]byte
	copy(digest[:], bytes)
	return systemTransferCatalogPin{versionID: version, sha256: digest}, nil
}

func systemTransferPublicPin(
	catalog publicresource.FieldCatalog,
) (systemTransferCatalogPin, error) {
	_, version, digest, err := catalog.Reference()
	if err != nil {
		return systemTransferCatalogPin{}, ErrSystemTransferUnavailable
	}
	parsed, err := uuid.Parse(version.String())
	if err != nil || parsed == uuid.Nil {
		return systemTransferCatalogPin{}, ErrSystemTransferUnavailable
	}
	return systemTransferCatalogPin{versionID: parsed, sha256: digest}, nil
}

func (pin systemTransferCatalogPin) same(other systemTransferCatalogPin) bool {
	return pin.versionID == other.versionID && pin.versionID != uuid.Nil &&
		subtle.ConstantTimeCompare(pin.sha256[:], other.sha256[:]) == 1
}
