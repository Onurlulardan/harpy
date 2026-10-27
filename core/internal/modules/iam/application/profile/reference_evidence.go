package profile

import "github.com/google/uuid"

type RetirementTarget struct {
	Resource              ResourceKind
	FieldID               uuid.UUID
	ExpectedFieldRevision int64
}

func (target RetirementTarget) Validate() error {
	if !target.Resource.Valid() || !stableID(target.FieldID) ||
		target.ExpectedFieldRevision < 1 {
		return ErrInvalidCommand
	}
	return nil
}

type ReferenceEvidence struct {
	Target         RetirementTarget
	ManifestSHA256 [32]byte
	CheckedOwners  int32
	LiveReferences int64
}

func (evidence ReferenceEvidence) Validate(target RetirementTarget) error {
	if evidence.Target != target || evidence.ManifestSHA256 == [32]byte{} ||
		evidence.CheckedOwners < 1 || evidence.LiveReferences < 0 {
		return ErrEvidenceMissing
	}
	if evidence.LiveReferences != 0 {
		return ErrReferencePresent
	}
	return nil
}
