package artifact

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

var proposalPurposes = map[string]bool{
	"initial_environment_authority": true,
	"initial_tenant_authority":      true,
	"owner_transfer":                true,
	"protected_admin_change":        true,
	"recovery_anchor_change":        true,
	"proposal_signer_rotation":      true,
	"compensate_uncommitted_ack":    true,
}

func validProposal(value Proposal, policy ExpiryPolicy) bool {
	issued, issuedOK := parseInstant(value.IssuedAt)
	expires, expiresOK := parseInstant(value.ExpiresAt)
	duration := 10 * time.Minute
	if policy == ExpiryAirGap {
		duration = 7 * 24 * time.Hour
	} else if policy != ExpiryOnline {
		return false
	}
	if value.Schema != ProposalSchema || !proposalPurposes[value.Purpose] ||
		!validUUID(value.OperationID) || !validUUID(value.EnvironmentID) ||
		!validUUID(value.RegistryBindingID) || !validUUID(value.ScopeID) ||
		!validDigest(value.StateSHA256) || !validOptionalDigest(value.PreviousHeadSHA256) ||
		!validOptionalDigest(value.RecoveryTrustBundleSHA256) ||
		!validOptionalDigest(value.CompensatesAckSHA256) || !validKID(value.SignerKID) ||
		!issuedOK || !expiresOK || expires.Sub(issued) != duration {
		return false
	}
	if value.DeploymentMode != ModeOnPremise && value.DeploymentMode != ModeSaaS {
		return false
	}
	if value.ScopeKind != ScopeEnvironment && value.ScopeKind != ScopeTenant {
		return false
	}
	if value.ScopeKind == ScopeEnvironment && value.ScopeID != value.EnvironmentID {
		return false
	}
	if value.Generation < 1 || (value.Generation == 1) != (value.PreviousHeadSHA256 == nil) {
		return false
	}
	if !validRotationEvidence(value) {
		return false
	}
	isCompensation := value.Purpose == "compensate_uncommitted_ack"
	return isCompensation == (value.CompensatesAckSHA256 != nil)
}

func validRotationEvidence(value Proposal) bool {
	isRotation := value.Purpose == "proposal_signer_rotation"
	if !isRotation {
		return value.SignerRotation == nil
	}
	rotation := value.SignerRotation
	if rotation == nil || rotation.PreviousSignerGeneration < 1 ||
		!validKID(rotation.NewSignerKID) || !validDigest(rotation.NewSignerPublicKeySHA256) {
		return false
	}
	if value.DeploymentMode == ModeSaaS {
		return rotation.RecoveryTrustBundleBytes == nil &&
			value.RecoveryTrustBundleSHA256 == nil
	}
	if rotation.RecoveryTrustBundleBytes == nil || value.RecoveryTrustBundleSHA256 == nil {
		return false
	}
	document, err := base64.RawURLEncoding.DecodeString(*rotation.RecoveryTrustBundleBytes)
	if err != nil || len(document) == 0 || len(document) > 32*1024 ||
		base64.RawURLEncoding.EncodeToString(document) != *rotation.RecoveryTrustBundleBytes {
		return false
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:]) == *value.RecoveryTrustBundleSHA256
}
