package attestation

import cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"

type VerifiedCutoverTerminalEvidence struct {
	evidence  cutoverv1.TerminalEvidence
	digest    string
	canonical []byte
	verified  bool
}

func (value VerifiedCutoverTerminalEvidence) SHA256() string { return value.digest }

func (value VerifiedCutoverTerminalEvidence) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}

func (value VerifiedCutoverTerminalEvidence) Value() cutoverv1.TerminalEvidence {
	result := value.evidence
	if value.evidence.FinalLSN != nil {
		copy := *value.evidence.FinalLSN
		result.FinalLSN = &copy
	}
	if value.evidence.InitialCopy != nil {
		copy := *value.evidence.InitialCopy
		result.InitialCopy = &copy
	}
	if value.evidence.CheckpointCount != nil {
		copy := *value.evidence.CheckpointCount
		result.CheckpointCount = &copy
	}
	result.Checkpoints = append([]cutoverv1.TerminalCheckpointEvidence(nil),
		value.evidence.Checkpoints...)
	if value.evidence.JournalTransactionCount != nil {
		copy := *value.evidence.JournalTransactionCount
		result.JournalTransactionCount = &copy
	}
	if value.evidence.AppliedChangeCount != nil {
		copy := *value.evidence.AppliedChangeCount
		result.AppliedChangeCount = &copy
	}
	if value.evidence.JournalCoverageRootSHA256 != nil {
		copy := *value.evidence.JournalCoverageRootSHA256
		result.JournalCoverageRootSHA256 = &copy
	}
	if value.evidence.OutcomeCode != nil {
		copy := *value.evidence.OutcomeCode
		result.OutcomeCode = &copy
	}
	if value.evidence.OutcomeEvidenceSHA256 != nil {
		copy := *value.evidence.OutcomeEvidenceSHA256
		result.OutcomeEvidenceSHA256 = &copy
	}
	return result
}
