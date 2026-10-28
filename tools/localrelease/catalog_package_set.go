package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/tools/internal/firstpartycatalog"
)

func buildCatalogPackageSet(
	releaseID string,
	authority catalogPackageAuthority,
) (artifact, error) {
	references, err := firstpartycatalog.BuildAll()
	if err != nil {
		return artifact{}, err
	}
	sort.Slice(references, func(left, right int) bool {
		return catalogPackageLess(references[left], references[right])
	})
	set := catalogpackagev1.ReleaseSet{
		ContractVersion: catalogpackagev1.ContractVersion,
		ReleaseID:       releaseID,
		Packages:        make([]catalogpackagev1.SealedPackage, 0, len(references)),
	}
	for _, reference := range references {
		sealed, err := sealCatalogPackage(reference, authority)
		if err != nil {
			return artifact{}, err
		}
		set.Packages = append(set.Packages, sealed)
	}
	document, err := json.Marshal(set)
	if err != nil {
		return artifact{}, err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, catalogpackagev1.MaximumPackageSetBytes,
	)
	if err != nil {
		return artifact{}, err
	}
	if _, err := catalogpackagev1.DecodeCanonical(canonical); err != nil {
		return artifact{}, err
	}
	digest := sha256.Sum256(canonical)
	return artifact{
		name:     "catalog-package-set.json",
		digest:   hex.EncodeToString(digest[:]),
		document: canonical,
	}, nil
}

func catalogPackageLess(
	left, right firstpartycatalog.Package,
) bool {
	if left.PackageID != right.PackageID {
		return left.PackageID < right.PackageID
	}
	leftVersion, leftOK := catalogSemver(left.Version)
	rightVersion, rightOK := catalogSemver(right.Version)
	if !leftOK || !rightOK {
		return left.Version < right.Version
	}
	for index := range leftVersion {
		if leftVersion[index] != rightVersion[index] {
			return leftVersion[index] < rightVersion[index]
		}
	}
	return false
}

func catalogSemver(value string) ([3]uint64, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return [3]uint64{}, false
	}
	var result [3]uint64
	for index, part := range parts {
		parsed, err := strconv.ParseUint(part, 10, 64)
		if err != nil || strconv.FormatUint(parsed, 10) != part {
			return [3]uint64{}, false
		}
		result[index] = parsed
	}
	return result, true
}

func sealCatalogPackage(
	reference firstpartycatalog.Package,
	authority catalogPackageAuthority,
) (catalogpackagev1.SealedPackage, error) {
	payloads := append(
		[]catalogpackagev1.OwnerPayload(nil),
		reference.OwnerPayloads...,
	)
	for index := range payloads {
		payloads[index].Document = append([]byte(nil), payloads[index].Document...)
	}
	sort.Slice(payloads, func(left, right int) bool {
		return payloads[left].ComponentID < payloads[right].ComponentID
	})
	signature, err := authority.Sign(reference.Manifest)
	if err != nil {
		return catalogpackagev1.SealedPackage{}, err
	}
	return catalogpackagev1.SealedPackage{
		Manifest:       append([]byte(nil), reference.Manifest...),
		ManifestSHA256: reference.ManifestSHA256,
		Signature:      base64.RawURLEncoding.EncodeToString(signature),
		KeyID:          authority.keyID,
		OwnerPayloads:  payloads,
	}, nil
}
