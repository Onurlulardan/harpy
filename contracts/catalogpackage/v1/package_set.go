package catalogpackagev1

import "encoding/json"

const (
	ContractVersion        = "catalog-release-package-set.v1"
	MaximumPackageSetBytes = 16 << 20
	MaximumManifestBytes   = 1 << 20
	MaximumPayloadBytes    = 4 << 20
)

type ReleaseSet struct {
	ContractVersion string          `json:"contract_version"`
	ReleaseID       string          `json:"release_id"`
	Packages        []SealedPackage `json:"packages"`
}

type SealedPackage struct {
	Manifest       json.RawMessage `json:"manifest"`
	ManifestSHA256 string          `json:"manifest_sha256"`
	Signature      string          `json:"signature"`
	KeyID          string          `json:"key_id"`
	OwnerPayloads  []OwnerPayload  `json:"owner_payloads"`
}

type OwnerPayload struct {
	ComponentID   string          `json:"component_id"`
	Owner         string          `json:"owner"`
	Contract      string          `json:"contract"`
	ContentSHA256 string          `json:"content_sha256"`
	Document      json.RawMessage `json:"document"`
}

func (set ReleaseSet) Clone() ReleaseSet {
	cloned := set
	cloned.Packages = make([]SealedPackage, len(set.Packages))
	for index := range set.Packages {
		cloned.Packages[index] = set.Packages[index].clone()
	}
	return cloned
}

func (value SealedPackage) clone() SealedPackage {
	cloned := value
	cloned.Manifest = append(json.RawMessage(nil), value.Manifest...)
	cloned.OwnerPayloads = make([]OwnerPayload, len(value.OwnerPayloads))
	for index := range value.OwnerPayloads {
		cloned.OwnerPayloads[index] = value.OwnerPayloads[index]
		cloned.OwnerPayloads[index].Document = append(
			json.RawMessage(nil), value.OwnerPayloads[index].Document...,
		)
	}
	return cloned
}
