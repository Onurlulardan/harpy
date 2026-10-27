package packageregistry

import "time"

import catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"

type SealedVersion struct {
	PackageID      string
	Slug           string
	Version        string
	Manifest       []byte
	ManifestSHA256 string
	Signature      []byte
	KeyID          string
	ReleaseID      string
	RegisteredAt   time.Time
	OwnerPayloads  []catalogpackagev1.OwnerPayload
}

func (value SealedVersion) Clone() SealedVersion {
	value.Manifest = append([]byte(nil), value.Manifest...)
	value.Signature = append([]byte(nil), value.Signature...)
	value.OwnerPayloads = cloneOwnerPayloads(value.OwnerPayloads)
	return value
}

func cloneOwnerPayloads(
	payloads []catalogpackagev1.OwnerPayload,
) []catalogpackagev1.OwnerPayload {
	cloned := make([]catalogpackagev1.OwnerPayload, len(payloads))
	for index := range payloads {
		cloned[index] = payloads[index]
		cloned[index].Document = append([]byte(nil), payloads[index].Document...)
	}
	return cloned
}

type Package struct {
	ID        string
	Slug      string
	CreatedAt time.Time
}

type DeclaredDependency struct {
	PackageID    string
	VersionRange string
	Kind         string
}
