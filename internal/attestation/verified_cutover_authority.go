package attestation

import cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"

type VerifiedAuthorityAdoptionOperation struct {
	operation cutoverv1.AuthorityAdoptionOperation
	digest    string
	canonical []byte
	verified  bool
}

func newVerifiedAuthorityAdoptionOperation(
	operation cutoverv1.AuthorityAdoptionOperation, digest string, canonical []byte,
) VerifiedAuthorityAdoptionOperation {
	return VerifiedAuthorityAdoptionOperation{
		operation: cloneAuthorityOperation(operation), digest: digest,
		canonical: append([]byte(nil), canonical...), verified: true,
	}
}

func (value VerifiedAuthorityAdoptionOperation) Value() cutoverv1.AuthorityAdoptionOperation {
	return cloneAuthorityOperation(value.operation)
}
func (value VerifiedAuthorityAdoptionOperation) ManifestSHA256() string { return value.digest }
func (value VerifiedAuthorityAdoptionOperation) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}

type VerifiedAuthorityAdoptionReceipt struct {
	receipt   cutoverv1.AuthorityAdoptionReceipt
	digest    string
	canonical []byte
	verified  bool
}

func newVerifiedAuthorityAdoptionReceipt(
	receipt cutoverv1.AuthorityAdoptionReceipt, digest string, canonical []byte,
) VerifiedAuthorityAdoptionReceipt {
	return VerifiedAuthorityAdoptionReceipt{
		receipt: cloneAuthorityReceipt(receipt), digest: digest,
		canonical: append([]byte(nil), canonical...), verified: true,
	}
}

func (value VerifiedAuthorityAdoptionReceipt) Value() cutoverv1.AuthorityAdoptionReceipt {
	return cloneAuthorityReceipt(value.receipt)
}
func (value VerifiedAuthorityAdoptionReceipt) ReceiptSHA256() string { return value.digest }
func (value VerifiedAuthorityAdoptionReceipt) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}

func cloneAuthorityOperation(
	value cutoverv1.AuthorityAdoptionOperation,
) cutoverv1.AuthorityAdoptionOperation {
	value.PreviousAuthorityReceiptSHA256 = cloneStringPointer(value.PreviousAuthorityReceiptSHA256)
	value.UpgradeArtifactSHA256 = cloneStringPointer(value.UpgradeArtifactSHA256)
	value.Objects = append([]cutoverv1.AuthorityAdoptionObject(nil), value.Objects...)
	value.Dependencies = append(
		[]cutoverv1.AuthorityCatalogDependency(nil), value.Dependencies...,
	)
	value.QuiescencePolicy.Roles = append(
		[]cutoverv1.AuthorityQuiescedRolePolicy(nil), value.QuiescencePolicy.Roles...,
	)
	return value
}

func cloneAuthorityReceipt(
	value cutoverv1.AuthorityAdoptionReceipt,
) cutoverv1.AuthorityAdoptionReceipt {
	value.PreviousAuthorityReceiptSHA256 = cloneStringPointer(value.PreviousAuthorityReceiptSHA256)
	value.UpgradeArtifactSHA256 = cloneStringPointer(value.UpgradeArtifactSHA256)
	value.Objects = append([]cutoverv1.AuthorityAdoptionReceiptObject(nil), value.Objects...)
	value.Dependencies = append(
		[]cutoverv1.AuthorityCatalogDependencyReceipt(nil), value.Dependencies...,
	)
	value.QuiescenceEvidence.Roles = append(
		[]cutoverv1.AuthorityQuiescedRoleEvidence(nil), value.QuiescenceEvidence.Roles...,
	)
	if value.PreviousCatalogCheckpoint != nil {
		checkpoint := *value.PreviousCatalogCheckpoint
		checkpoint.Objects = append(
			[]cutoverv1.AuthorityAdoptionReceiptObject(nil), checkpoint.Objects...,
		)
		value.PreviousCatalogCheckpoint = &checkpoint
	}
	return value
}
