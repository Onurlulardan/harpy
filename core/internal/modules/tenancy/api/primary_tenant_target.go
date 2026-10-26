package api

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type PrimaryTenantPurpose string

const (
	PrimaryTenantPurposeOnPremise PrimaryTenantPurpose = "onpremise_primary"
	PrimaryTenantPurposeSaaS      PrimaryTenantPurpose = "saas_control"
)

type PrimaryTenantTarget struct {
	environmentID string
	mode          DeploymentMode
	tenantID      string
	purpose       PrimaryTenantPurpose
	canonicalSlug string
}

func NewPrimaryTenantTarget(
	environmentID string,
	mode DeploymentMode,
	tenantID string,
	purpose PrimaryTenantPurpose,
	canonicalSlug string,
) (PrimaryTenantTarget, error) {
	if !canonicalUUID(environmentID) || !canonicalUUID(tenantID) ||
		!validCanonicalTenantSlug(canonicalSlug) ||
		!validPrimaryPurpose(mode, purpose) {
		return PrimaryTenantTarget{}, ErrPrimaryTenantTargetInvalid
	}
	return PrimaryTenantTarget{
		environmentID: environmentID,
		mode:          mode,
		tenantID:      tenantID,
		purpose:       purpose,
		canonicalSlug: canonicalSlug,
	}, nil
}

func (target PrimaryTenantTarget) EnvironmentID() string {
	return target.environmentID
}

func (target PrimaryTenantTarget) DeploymentMode() DeploymentMode {
	return target.mode
}

func (target PrimaryTenantTarget) TenantID() string {
	return target.tenantID
}

func (target PrimaryTenantTarget) Purpose() PrimaryTenantPurpose {
	return target.purpose
}

func (target PrimaryTenantTarget) CanonicalSlug() string {
	return target.canonicalSlug
}

func (target PrimaryTenantTarget) PublicTenantSlug() (string, bool) {
	if target.mode != DeploymentModeOnPremise ||
		target.purpose != PrimaryTenantPurposeOnPremise {
		return "", false
	}
	return target.canonicalSlug, true
}

func (target PrimaryTenantTarget) IsValid() bool {
	_, err := NewPrimaryTenantTarget(
		target.environmentID,
		target.mode,
		target.tenantID,
		target.purpose,
		target.canonicalSlug,
	)
	return err == nil
}

func validPrimaryPurpose(
	mode DeploymentMode,
	purpose PrimaryTenantPurpose,
) bool {
	return mode == DeploymentModeOnPremise &&
		purpose == PrimaryTenantPurposeOnPremise ||
		mode == DeploymentModeSaaS && purpose == PrimaryTenantPurposeSaaS
}

func validCanonicalTenantSlug(slug string) bool {
	if len(slug) == 0 || len(slug) > 100 || !utf8.ValidString(slug) ||
		strings.TrimSpace(slug) != slug {
		return false
	}
	for _, character := range slug {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
