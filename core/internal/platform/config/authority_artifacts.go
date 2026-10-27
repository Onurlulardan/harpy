package config

import "github.com/google/uuid"

type AuthorityArtifacts struct {
	manifestPath  AbsolutePath
	signaturePath AbsolutePath
}

func loadAuthorityArtifacts(values environmentValues) (AuthorityArtifacts, error) {
	manifest, err := loadAbsolutePath(values, "HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH")
	if err != nil {
		return AuthorityArtifacts{}, err
	}
	signature, err := loadAbsolutePath(values, "HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH")
	if err != nil {
		return AuthorityArtifacts{}, err
	}
	if duplicateReleasePath(manifest, signature) {
		return AuthorityArtifacts{}, configError(
			"HARPY_BOOTSTRAP_AUTHORITY_*_PATH", "authority paths must be distinct",
		)
	}
	return AuthorityArtifacts{manifestPath: manifest, signaturePath: signature}, nil
}

func (artifacts AuthorityArtifacts) ManifestPath() AbsolutePath {
	return artifacts.manifestPath
}

func (artifacts AuthorityArtifacts) SignaturePath() AbsolutePath {
	return artifacts.signaturePath
}

type PrivilegedAuthority struct {
	artifacts AuthorityArtifacts
	leaseID   string
}

func loadPrivilegedAuthority(values environmentValues) (PrivilegedAuthority, error) {
	artifacts, err := loadAuthorityArtifacts(values)
	if err != nil {
		return PrivilegedAuthority{}, err
	}
	leaseID, found := values.lookup("HARPY_CREDENTIAL_LEASE_ID")
	if !found {
		return PrivilegedAuthority{}, configError("HARPY_CREDENTIAL_LEASE_ID", "is required")
	}
	parsed, err := uuid.Parse(leaseID)
	if err != nil || parsed == uuid.Nil || parsed.String() != leaseID {
		return PrivilegedAuthority{}, configError(
			"HARPY_CREDENTIAL_LEASE_ID", "must be a canonical non-zero UUID",
		)
	}
	return PrivilegedAuthority{artifacts: artifacts, leaseID: leaseID}, nil
}

func (authority PrivilegedAuthority) Artifacts() AuthorityArtifacts {
	return authority.artifacts
}

func (authority PrivilegedAuthority) LeaseID() string { return authority.leaseID }
