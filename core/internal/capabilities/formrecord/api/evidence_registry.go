package recordkernel

import "slices"

func digestRegistry(snapshot RegistrySnapshot) ([32]byte, error) {
	ordered := append([]Descriptor(nil), snapshot.Types...)
	slices.SortFunc(ordered, func(left, right Descriptor) int {
		return compareText(string(left.TypeID), string(right.TypeID))
	})
	documents := make([]evidenceDescriptor, len(ordered))
	for index, descriptor := range ordered {
		if descriptor.TypeID == "" || index > 0 && descriptor.TypeID == ordered[index-1].TypeID {
			return [32]byte{}, ErrCatalogInvalid
		}
		documents[index] = descriptorEvidence(descriptor)
	}
	return evidenceDigest(struct {
		ContractVersion string               `json:"contract_version"`
		HardLimits      evidenceLimits       `json:"hard_limits"`
		Types           []evidenceDescriptor `json:"types"`
	}{snapshot.ContractVersion, limitsEvidence(snapshot.HardLimits), documents})
}

func digestFields(fields []PublishedField) ([32]byte, error) {
	ordered := append([]PublishedField(nil), fields...)
	slices.SortFunc(ordered, func(left, right PublishedField) int {
		return slices.Compare(left.FieldID[:], right.FieldID[:])
	})
	documents := make([]EvidenceField, len(ordered))
	for index, field := range ordered {
		if field.FieldID == [16]byte{} || field.PublicName == "" || field.Type == "" ||
			field.Role == "" || field.Generation < 1 ||
			index > 0 && field.FieldID == ordered[index-1].FieldID {
			return [32]byte{}, ErrCatalogInvalid
		}
		documents[index] = EvidenceField{field.FieldID.String(), field.PublicName,
			field.Type, field.Role, field.Generation}
	}
	return evidenceDigest(struct {
		Contract string          `json:"contract"`
		Fields   []EvidenceField `json:"fields"`
	}{"forms-publication-field-catalog.v1", documents})
}
