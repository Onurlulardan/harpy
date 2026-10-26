package maintenance

type Kind string

const (
	KindDirectoryRepair         Kind = "iam.directory.repair"
	KindDirectoryChecksum       Kind = "iam.directory.checksum"
	KindCredentialBcryptRetire  Kind = "iam.credential.bcrypt-retire"
	KindIdentifierRetire        Kind = "iam.identifier.retire"
	KindDirectoryTombstonePurge Kind = "iam.directory.tombstone-purge"
	KindSecurityArtifactPurge   Kind = "iam.security-artifact-purge"
	KindProfileBackfill         Kind = "iam.profile.backfill"
)

const PayloadRevision = 1

func (kind Kind) Valid() bool {
	switch kind {
	case KindDirectoryRepair, KindDirectoryChecksum,
		KindCredentialBcryptRetire, KindIdentifierRetire,
		KindDirectoryTombstonePurge, KindSecurityArtifactPurge,
		KindProfileBackfill:
		return true
	default:
		return false
	}
}

func (kind Kind) Recurring() bool {
	return kind.Valid() && kind != KindProfileBackfill
}
