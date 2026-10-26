package recoverystorage

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	regionPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	drivePath     = regexp.MustCompile(`^[A-Za-z]:[\\/]`)
	uuidPattern   = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
)

func (manifest Manifest) Valid() bool {
	recovery := StorageProfile{
		Endpoint: manifest.Endpoint, Region: manifest.Region, Bucket: manifest.Bucket,
		ExpectedOwner: manifest.ExpectedOwner, PathStyle: manifest.PathStyle,
		TLSCASHA256: manifest.TLSCASHA256,
	}
	endpoint, valid := validStorage(recovery)
	if !valid {
		return false
	}
	if manifest.Schema != Schema || manifest.DeploymentMode != "onpremise" ||
		manifest.Bucket != Bucket ||
		manifest.DocumentsSource.Bucket == manifest.Bucket ||
		manifest.DocumentsRehydrateTarget.Bucket == manifest.Bucket ||
		manifest.DocumentsSource.Bucket == manifest.DocumentsRehydrateTarget.Bucket ||
		!storageProfileValid(manifest.DocumentsSource) ||
		!storageProfileValid(manifest.DocumentsRehydrateTarget) ||
		!validProtection(manifest, endpoint.Scheme) || !validObjectLock(manifest.ObjectLock) ||
		!validProvider(manifest.RecoveryKeyProvider) {
		return false
	}
	paths := []string{
		manifest.Credentials.RecoveryWrite, manifest.Credentials.RetentionDelete,
		manifest.Credentials.RestoreRead, manifest.Credentials.DocumentsSourceRead,
		manifest.Credentials.DocumentsRehydrate}
	seen := make(map[string]bool, len(paths))
	for _, path := range paths {
		if !absolutePath(path) || seen[path] {
			return false
		}
		seen[path] = true
	}
	return true
}

func validStorage(profile StorageProfile) (*url.URL, bool) {
	endpoint, err := url.Parse(profile.Endpoint)
	valid := err == nil && endpoint != nil && endpoint.User == nil &&
		endpoint.Host != "" &&
		(endpoint.Scheme == "https" || endpoint.Scheme == "http") &&
		endpoint.Path == "" && endpoint.RawQuery == "" && endpoint.Fragment == "" &&
		endpoint.String() == profile.Endpoint
	return endpoint, valid
}

func storageProfileValid(profile StorageProfile) bool {
	_, valid := validStorage(profile)
	return valid && profile.Bucket != "" && profile.ExpectedOwner != "" &&
		regionPattern.MatchString(profile.Region) &&
		digestPattern.MatchString(profile.TLSCASHA256)
}

// AllowsProtected reports whether this storage target may produce a protected
// recovery set in the named runtime environment. Local-only storage is useful
// for development rehearsal, but it is never DR evidence in staging or
// production.
func (manifest Manifest) AllowsProtected(environment string) bool {
	if !manifest.Valid() {
		return false
	}
	switch environment {
	case "development":
		return manifest.ProtectionClass == "local_only" ||
			manifest.ProtectionClass == "offsite"
	case "staging", "production":
		return manifest.ProtectionClass == "offsite"
	default:
		return false
	}
}

func validProtection(manifest Manifest, scheme string) bool {
	switch manifest.ProtectionClass {
	case "offsite":
		return scheme == "https" && digestPattern.MatchString(manifest.OffsiteProofSHA256)
	case "local_only":
		return manifest.OffsiteProofSHA256 == ""
	default:
		return false
	}
}

func validObjectLock(policy ObjectLockPolicy) bool {
	return policy.Enabled && policy.Mode == "COMPLIANCE" &&
		policy.RetentionDays >= 1 && policy.RetentionDays <= 36500
}

func validProvider(provider KeyProvider) bool {
	return (provider.Kind == "local_keyring" || provider.Kind == "forgeguard") &&
		len(provider.Handle) >= 1 && len(provider.Handle) <= 512 &&
		!strings.ContainsAny(provider.Handle, "\x00\r\n")
}

func absolutePath(value string) bool {
	return strings.HasPrefix(value, "/") || drivePath.MatchString(value)
}

func validUUID(value string) bool {
	return uuidPattern.MatchString(value)
}
