package attestation

import catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"

// VerifiedCatalogPackageSet is an opaque release-bound package publication set.
type VerifiedCatalogPackageSet struct {
	value    catalogpackagev1.ReleaseSet
	verified bool
}

func newVerifiedCatalogPackageSet(
	value catalogpackagev1.ReleaseSet,
) VerifiedCatalogPackageSet {
	return VerifiedCatalogPackageSet{value: value.Clone(), verified: true}
}

func (set VerifiedCatalogPackageSet) Value() catalogpackagev1.ReleaseSet {
	if !set.verified {
		return catalogpackagev1.ReleaseSet{}
	}
	return set.value.Clone()
}

func (set VerifiedCatalogPackageSet) IsVerified() bool {
	return set.verified
}
