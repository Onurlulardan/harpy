package recordkernel

import "github.com/google/uuid"

type Projection struct {
	originalContract       string
	canonicalDefinition    []byte
	definitionSHA256       [32]byte
	storageProjection      []byte
	storageSHA256          [32]byte
	presentationProjection []byte
	presentationSHA256     [32]byte
	recordRules            RecordRules
	recordRulesSHA256      [32]byte
	canonicalAttestation   []byte
	attestationSHA256      [32]byte
	relationTargetFormIDs  []uuid.UUID
}

func (projection Projection) OriginalContract() string {
	return projection.originalContract
}

func (projection Projection) CanonicalDefinition() []byte {
	return append([]byte(nil), projection.canonicalDefinition...)
}

func (projection Projection) DefinitionSHA256() [32]byte {
	return projection.definitionSHA256
}

func (projection Projection) StorageProjection() []byte {
	return append([]byte(nil), projection.storageProjection...)
}

func (projection Projection) StorageSHA256() [32]byte {
	return projection.storageSHA256
}

func (projection Projection) PresentationProjection() []byte {
	return append([]byte(nil), projection.presentationProjection...)
}

func (projection Projection) PresentationSHA256() [32]byte {
	return projection.presentationSHA256
}

func (projection Projection) RecordRules() RecordRules {
	return cloneRecordRules(projection.recordRules)
}

func (projection Projection) RecordRulesSHA256() [32]byte {
	return projection.recordRulesSHA256
}

func (projection Projection) CanonicalAttestation() []byte {
	return append([]byte(nil), projection.canonicalAttestation...)
}

func (projection Projection) AttestationSHA256() [32]byte {
	return projection.attestationSHA256
}

func (projection Projection) RelationTargetFormIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), projection.relationTargetFormIDs...)
}
