package recordkernel

import "crypto/sha256"

func buildProjection(
	originalContract string,
	canonicalDefinition []byte,
	projected []byte,
	rules RecordRules,
	recordBytes []byte,
) (Projection, error) {
	definitionSHA := sha256.Sum256(canonicalDefinition)
	projectionSHA := sha256.Sum256(projected)
	rulesSHA := sha256.Sum256(recordBytes)
	preimage := struct {
		Contract           string `json:"contract"`
		DefinitionContract string `json:"definition_contract"`
		DefinitionSHA256   string `json:"definition_sha256"`
		PresentationSHA256 string `json:"presentation_projection_sha256"`
		RecordContract     string `json:"record_contract"`
		RecordRulesSHA256  string `json:"record_rules_sha256"`
		StorageSHA256      string `json:"storage_projection_sha256"`
		Valid              bool   `json:"valid"`
	}{
		ProjectionContract, originalContract, hexDigest(definitionSHA),
		hexDigest(projectionSHA), RecordRulesContract, hexDigest(rulesSHA),
		hexDigest(projectionSHA), true,
	}
	attestation, err := canonicalObject(preimage, MaximumRecordBytes)
	if err != nil {
		return Projection{}, err
	}
	return Projection{
		originalContract:    originalContract,
		canonicalDefinition: append([]byte(nil), canonicalDefinition...),
		definitionSHA256:    definitionSHA,
		storageProjection:   append([]byte(nil), projected...), storageSHA256: projectionSHA,
		presentationProjection: append([]byte(nil), projected...),
		presentationSHA256:     projectionSHA, recordRules: cloneRecordRules(rules),
		recordRulesSHA256: rulesSHA, canonicalAttestation: attestation,
		attestationSHA256: sha256.Sum256(attestation),
	}, nil
}

func hexDigest(digest [32]byte) string {
	const digits = "0123456789abcdef"
	result := make([]byte, 64)
	for index, value := range digest {
		result[index*2] = digits[value>>4]
		result[index*2+1] = digits[value&15]
	}
	return string(result)
}
