package api

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type OnboardingPlan struct {
	input          OnboardingPlanInput
	stepPlanSHA256 string
}

func NewOnboardingPlan(input OnboardingPlanInput) (OnboardingPlan, error) {
	if !validOnboardingPlanInput(input) {
		return OnboardingPlan{}, ErrOnboardingInvalid
	}
	return OnboardingPlan{input: input, stepPlanSHA256: digestOnboardingPlan(input)}, nil
}

func validOnboardingPlanInput(input OnboardingPlanInput) bool {
	identities := []UUID{input.OperationID, input.EnvironmentID, input.TenantID,
		input.FirstUserID, input.RootOrganizationID, input.ProtectedAdminBindingID,
		input.Offering.VersionID, input.Authority.RegistryBindingID, input.Authority.ScopeID}
	for _, identity := range identities {
		if identity == NilUUID {
			return false
		}
	}
	if onboardingSchemaName(input.TenantID) != input.SchemaName ||
		!validOnboardingCombination(input.Kind, input.DeploymentMode, input.BindingPurpose) ||
		len(input.IdempotencyKey) < 1 || len(input.IdempotencyKey) > 128 ||
		strings.TrimSpace(input.IdempotencyKey) != input.IdempotencyKey ||
		!validOnboardingDigest(input.RequestSHA256) ||
		!validOnboardingDigest(input.ManifestSHA256) || !validOfferingPin(input.Offering) ||
		!validRecoveryPin(input) || !validAuthorityPlan(input) ||
		!microsecondUTC(input.CreatedAt) {
		return false
	}
	ownerRequired := input.Kind == OnboardingEnvironmentInitialize
	return ownerRequired && input.InstanceOwnerBindingID != NilUUID ||
		!ownerRequired && input.InstanceOwnerBindingID == NilUUID
}

func validOfferingPin(value BaseOfferingPin) bool {
	return value.VersionID != NilUUID && value.Revision > 0 &&
		validOnboardingDigest(value.SHA256) && microsecondUTC(value.EffectiveAt) &&
		(!value.Source.Present && value.Source.Sequence == 0 && value.Source.SHA256 == "" ||
			value.Source.Present && value.Source.Sequence > 0 &&
				validOnboardingDigest(value.Source.SHA256))
}

func validRecoveryPin(input OnboardingPlanInput) bool {
	value := input.RecoverySigning
	if input.DeploymentMode == DeploymentModeOnPremise {
		return value.Present && value.Generation > 0 &&
			validOnboardingDigest(value.TrustBundleSHA256) && input.Offering.Source.Present
	}
	return !value.Present && value.Generation == 0 && value.TrustBundleSHA256 == ""
}

func validAuthorityPlan(input OnboardingPlanInput) bool {
	value := input.Authority
	if input.Kind == OnboardingSaaSCustomerProvision {
		return value.ScopeKind == AuthorityScopeTenant && value.ScopeID == input.TenantID &&
			value.Generation == 1 && value.PreviousHeadSHA256 == "" &&
			validOnboardingDigest(value.StateSHA256) && validOnboardingDigest(value.ProposalSHA256)
	}
	if value.ScopeKind != AuthorityScopeEnvironment || value.ScopeID != input.EnvironmentID {
		return false
	}
	if input.DeploymentMode == DeploymentModeOnPremise {
		return value.Generation == 0 && value.PreviousHeadSHA256 == "" &&
			value.StateSHA256 == "" && value.ProposalSHA256 == ""
	}
	return value.Generation == 1 && value.PreviousHeadSHA256 == "" &&
		validOnboardingDigest(value.StateSHA256) && validOnboardingDigest(value.ProposalSHA256)
}

func microsecondUTC(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func digestOnboardingPlan(input OnboardingPlanInput) string {
	hash := sha256.New()
	values := []string{UUIDString(input.OperationID), UUIDString(input.EnvironmentID),
		UUIDString(input.TenantID), UUIDString(input.FirstUserID), UUIDString(input.RootOrganizationID),
		UUIDString(input.ProtectedAdminBindingID), UUIDString(input.InstanceOwnerBindingID),
		string(input.Kind), string(input.DeploymentMode), string(input.BindingPurpose),
		input.SchemaName, input.IdempotencyKey, input.RequestSHA256, input.ManifestSHA256,
		UUIDString(input.Offering.VersionID), strconv.FormatInt(input.Offering.Revision, 10),
		input.Offering.SHA256, input.Offering.EffectiveAt.Format(time.RFC3339Nano),
		strconv.FormatInt(input.Offering.Source.Sequence, 10), input.Offering.Source.SHA256,
		strconv.FormatBool(input.Offering.Source.Present),
		strconv.FormatInt(input.RecoverySigning.Generation, 10), input.RecoverySigning.TrustBundleSHA256,
		strconv.FormatBool(input.RecoverySigning.Present),
		UUIDString(input.Authority.RegistryBindingID), string(input.Authority.ScopeKind),
		UUIDString(input.Authority.ScopeID), strconv.FormatInt(input.Authority.Generation, 10),
		input.Authority.PreviousHeadSHA256, input.Authority.StateSHA256,
		input.Authority.ProposalSHA256, input.CreatedAt.Format(time.RFC3339Nano),
		activationManifestText(input.Kind)}
	for _, value := range values {
		_, _ = hash.Write([]byte(strconv.Itoa(len(value))))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
