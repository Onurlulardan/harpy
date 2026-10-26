package transaction

import "github.com/google/uuid"

func validCredentialSource(evidence CredentialTargetEvidence) bool {
	switch evidence.LocatorKind {
	case CredentialLoginDirectory:
		return evidence.SourceUserID != uuid.Nil &&
			evidence.SourceIdentifierID != uuid.Nil &&
			evidence.SourceIdentifierKind.valid() &&
			emptyRefreshSource(evidence) && emptyGrantSource(evidence)
	case CredentialRefreshLocator:
		return evidence.SourceTokenID != uuid.Nil &&
			evidence.SourceFamilyID != uuid.Nil &&
			evidence.SourceSelector != [16]byte{} &&
			emptyLoginSource(evidence) && emptyGrantSource(evidence)
	default:
		return emptyLoginSource(evidence) && emptyRefreshSource(evidence) &&
			emptyGrantSource(evidence)
	}
}

func (target VerifiedCredentialTarget) credentialSourceValid() bool {
	return validCredentialSource(CredentialTargetEvidence{
		LocatorKind:          target.locatorKind,
		SourceUserID:         target.sourceUserID,
		SourceIdentifierID:   target.sourceIdentifierID,
		SourceIdentifierKind: target.sourceIdentifierKind,
		SourceTokenID:        target.sourceTokenID,
		SourceFamilyID:       target.sourceFamilyID,
		SourceRequestID:      target.sourceRequestID,
		SourceGrantID:        target.sourceGrantID,
		SourceSelector:       target.sourceSelector,
	})
}

func (target VerifiedCredentialTarget) matchesLookup(lookup CredentialLookup) bool {
	if target.locatorKind != lookup.locatorKind {
		return false
	}
	switch target.locatorKind {
	case CredentialLoginDirectory:
		return target.sourceIdentifierKind == lookup.identifier
	case CredentialRefreshLocator:
		return target.sourceSelector == [16]byte(lookup.key[:16])
	default:
		return true
	}
}

func emptyLoginSource(evidence CredentialTargetEvidence) bool {
	return evidence.SourceUserID == uuid.Nil && evidence.SourceIdentifierID == uuid.Nil &&
		evidence.SourceIdentifierKind == ""
}

func emptyRefreshSource(evidence CredentialTargetEvidence) bool {
	return evidence.SourceTokenID == uuid.Nil && evidence.SourceFamilyID == uuid.Nil &&
		evidence.SourceSelector == [16]byte{}
}

func emptyGrantSource(evidence CredentialTargetEvidence) bool {
	return evidence.SourceRequestID == uuid.Nil && evidence.SourceGrantID == uuid.Nil
}
