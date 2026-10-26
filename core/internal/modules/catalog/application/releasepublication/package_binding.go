package releasepublication

import (
	"encoding/base64"
	"fmt"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	analyticscatalogpayload "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/catalogpayload"
	democatalogpayload "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/catalogpayload"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	iamcatalogpayload "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/catalogpayload"
)

func PackageVersionsFromReleaseSet(
	set catalogpackagev1.ReleaseSet,
) ([]packageregistry.SealedVersion, error) {
	if len(set.Packages) == 0 {
		return nil, fmt.Errorf("catalog package set binding failed")
	}
	packageIDs := make(map[string]bool, len(set.Packages))
	slugs := make(map[string]bool, len(set.Packages))
	versions := make(map[string]bool, len(set.Packages))
	result := make([]packageregistry.SealedVersion, 0, len(set.Packages))
	for _, sealed := range set.Packages {
		manifest, err := packagecontract.Decode(sealed.Manifest)
		identity := manifest.PackageID.String() + "@" + manifest.Version.String()
		if err != nil || packageIDs[manifest.PackageID.String()] ||
			slugs[manifest.Slug] || versions[identity] {
			return nil, fmt.Errorf("catalog package identity binding failed")
		}
		packageIDs[manifest.PackageID.String()] = true
		slugs[manifest.Slug] = true
		versions[identity] = true
		if err := verifyPackagePayloads(manifest, sealed.OwnerPayloads); err != nil {
			return nil, err
		}
		signature, err := base64.RawURLEncoding.DecodeString(sealed.Signature)
		if err != nil {
			return nil, fmt.Errorf("catalog package signature binding failed")
		}
		result = append(result, packageregistry.SealedVersion{
			PackageID: manifest.PackageID.String(), Slug: manifest.Slug,
			Version:        manifest.Version.String(),
			Manifest:       append([]byte(nil), sealed.Manifest...),
			ManifestSHA256: sealed.ManifestSHA256,
			Signature:      append([]byte(nil), signature...), KeyID: sealed.KeyID,
			OwnerPayloads: clonePackagePayloads(sealed.OwnerPayloads),
		})
	}
	return result, nil
}

func verifyPackagePayloads(
	manifest packagecontract.Manifest,
	payloads []catalogpackagev1.OwnerPayload,
) error {
	if len(payloads) != len(manifest.Components) {
		return fmt.Errorf("catalog owner payload coverage differs")
	}
	byComponent := make(map[string]catalogpackagev1.OwnerPayload, len(payloads))
	for _, payload := range payloads {
		if _, duplicate := byComponent[payload.ComponentID]; duplicate {
			return fmt.Errorf("catalog owner payload identity is duplicated")
		}
		byComponent[payload.ComponentID] = payload
	}
	for _, component := range manifest.Components {
		payload, found := byComponent[component.ID.String()]
		if !found || payload.Owner != component.Owner ||
			payload.Contract != component.Contract ||
			payload.ContentSHA256 != component.ContentHash {
			return fmt.Errorf("catalog owner payload binding differs")
		}
		if err := verifyOwnerPayload(
			manifest.Slug, component.Owner, component.Contract, payload,
		); err != nil {
			return err
		}
		delete(byComponent, component.ID.String())
	}
	if len(byComponent) != 0 {
		return fmt.Errorf("catalog owner payload coverage differs")
	}
	return nil
}

func verifyOwnerPayload(
	packageSlug, owner, contract string,
	payload catalogpackagev1.OwnerPayload,
) error {
	var sourceContract string
	var verifyErr error
	switch {
	case owner == "iam" && contract == catalogpackagev1.ContractIAM:
		var verified iamcatalogpayload.Verified
		verified, verifyErr = iamcatalogpayload.Verify(
			payload.Document, payload.ComponentID, payload.ContentSHA256,
		)
		sourceContract = verified.Value().SourceContract
	case owner == "analytics" && contract == catalogpackagev1.ContractAnalytics:
		var verified analyticscatalogpayload.Verified
		verified, verifyErr = analyticscatalogpayload.Verify(
			payload.Document, payload.ComponentID, payload.ContentSHA256,
		)
		sourceContract = verified.Value().SourceContract
	case owner == "catalog" && contract == catalogpackagev1.ContractCatalog:
		var verified democatalogpayload.Verified
		verified, verifyErr = democatalogpayload.Verify(
			payload.Document, payload.ComponentID, payload.ContentSHA256,
		)
		sourceContract = verified.Value().SourceContract
	default:
		return fmt.Errorf("catalog owner payload contract is unsupported")
	}
	sourceSlug, sourceErr := catalogpackagev1.ParseFirstPartySourceContract(
		sourceContract,
	)
	if verifyErr != nil || sourceErr != nil || sourceSlug != packageSlug {
		return fmt.Errorf("%s catalog payload verification failed", owner)
	}
	return nil
}

func clonePackagePayloads(
	payloads []catalogpackagev1.OwnerPayload,
) []catalogpackagev1.OwnerPayload {
	cloned := make([]catalogpackagev1.OwnerPayload, len(payloads))
	for index := range payloads {
		cloned[index] = payloads[index]
		cloned[index].Document = append([]byte(nil), payloads[index].Document...)
	}
	return cloned
}
