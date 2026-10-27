package runtimev1

const (
	SourcePostgresEngineVersion = "15.18"
	SourcePostgresImageDigest   = "sha256:3073dc147f5b0ca05b36e10e04e73ca180b1e726e453c72a6fcc85f788afd0d1"
	SourcePostgresDataChecksums = false
	PostgresEngineVersion       = "18.4"
	PostgresServerVersionNumber = 180004
	PostgresImageDigest         = "sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"
	PostgresEncoding            = "UTF8"
	PostgresLocaleProvider      = "libc"
	PostgresCollation           = "en_US.utf8"
	PostgresCharacterType       = "en_US.utf8"
	PostgresDataChecksums       = true
	VectorExtensionVersion      = "0.8.3"
	DatabaseOwnerRole           = "harpy_owner"
)

type DatabaseOrigin string

const (
	DatabaseOriginFresh     DatabaseOrigin = "fresh"
	DatabaseOriginV1Adopted DatabaseOrigin = "v1_adopted"
)

func (origin DatabaseOrigin) Valid() bool {
	return origin == DatabaseOriginFresh || origin == DatabaseOriginV1Adopted
}
