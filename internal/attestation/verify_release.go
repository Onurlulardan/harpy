package attestation

import (
	"fmt"
	"regexp"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const maximumReleaseClockSkew = 5 * time.Minute

var (
	releaseIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	databasePattern  = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	digestPattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

func VerifyRelease(
	manifestDocument, signatureDocument []byte,
	root TrustRoot,
	now time.Time,
	targetDatabase string,
) (VerifiedRelease, error) {
	var manifest ReleaseManifest
	if err := decodeStrict(manifestDocument, &manifest); err != nil {
		return VerifiedRelease{}, fmt.Errorf("verify release manifest: %w", err)
	}
	issuedAt, err := validateReleaseManifest(manifest, now, targetDatabase)
	if err != nil {
		return VerifiedRelease{}, err
	}
	digest, err := verifyManifestSignature(
		manifestDocument, signatureDocument, "release-signature.v1", root, issuedAt,
	)
	if err != nil {
		return VerifiedRelease{}, fmt.Errorf("verify release: %w", err)
	}
	return newVerifiedRelease(manifest, digest), nil
}

func validateReleaseManifest(
	manifest ReleaseManifest,
	now time.Time,
	targetDatabase string,
) (time.Time, error) {
	issuedAt, issuedErr := parseUTCTimestamp(manifest.IssuedAt)
	expiresAt, expiresErr := parseUTCTimestamp(manifest.ExpiresAt)
	if issuedErr != nil || expiresErr != nil || !issuedAt.Before(expiresAt) ||
		issuedAt.After(now.Add(maximumReleaseClockSkew)) || !now.Before(expiresAt) {
		return time.Time{}, fmt.Errorf("release validity window is invalid")
	}
	if manifest.ContractVersion != "postgres-release-manifest.v1" ||
		!releaseIDPattern.MatchString(manifest.ReleaseID) ||
		!databasePattern.MatchString(manifest.DatabaseName) || manifest.DatabaseName != targetDatabase {
		return time.Time{}, fmt.Errorf("release identity is invalid")
	}
	origin := runtimev1.DatabaseOrigin(manifest.DatabaseOrigin)
	if !origin.Valid() || !validReleaseDatabaseContract(manifest) {
		return time.Time{}, fmt.Errorf("release database contract is invalid")
	}
	if manifest.RuntimeGeneration < 1 || manifest.MinimumDatabaseGeneration < 1 ||
		!validReleaseDigests(manifest) {
		return time.Time{}, fmt.Errorf("release generation or manifest digest is invalid")
	}
	return issuedAt, nil
}

func validReleaseDatabaseContract(manifest ReleaseManifest) bool {
	return manifest.DatabaseEncoding == runtimev1.PostgresEncoding &&
		manifest.DatabaseLocaleProvider == runtimev1.PostgresLocaleProvider &&
		manifest.DatabaseCollation == runtimev1.PostgresCollation &&
		manifest.DatabaseCType == runtimev1.PostgresCharacterType &&
		manifest.DatabaseOwner == runtimev1.DatabaseOwnerRole &&
		manifest.DataChecksums == runtimev1.PostgresDataChecksums &&
		manifest.PostgresImageDigest == runtimev1.PostgresImageDigest &&
		manifest.PostgresEngineVersion == runtimev1.PostgresEngineVersion &&
		manifest.VectorExtensionVersion == runtimev1.VectorExtensionVersion
}

func validReleaseDigests(manifest ReleaseManifest) bool {
	digests := []string{
		manifest.MigrationManifestSHA256, manifest.ExtensionManifestSHA256,
		manifest.RolesManifestSHA256, manifest.PrivilegesManifestSHA256,
		manifest.OwnershipManifestSHA256,
		manifest.BootstrapAuthoritySHA256, manifest.BootstrapOperationDigest,
		manifest.ExtensionOperationDigest, manifest.CredentialCleanupDigest,
		manifest.CutoverAuthorityOperationDigest, manifest.CatalogPackageSetSHA256,
	}
	for _, digest := range digests {
		if !digestPattern.MatchString(digest) {
			return false
		}
	}
	return true
}
