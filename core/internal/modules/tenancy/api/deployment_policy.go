package api

import (
	"strings"
	"time"
)

type DeploymentMode string

const (
	DeploymentModeOnPremise DeploymentMode = "onpremise"
	DeploymentModeSaaS      DeploymentMode = "saas"
)

type DeploymentPolicy struct {
	environmentID          string
	mode                   DeploymentMode
	maximumCustomerTenants int64
	hasCustomerLimit       bool
	createdAt              time.Time
}

func NewDeploymentPolicy(
	environmentID string,
	mode DeploymentMode,
	maximumCustomerTenants *int64,
	createdAt time.Time,
) (DeploymentPolicy, error) {
	if !canonicalUUID(environmentID) || createdAt.IsZero() ||
		createdAt.Location() != time.UTC || createdAt.Nanosecond()%1_000 != 0 {
		return DeploymentPolicy{}, ErrDeploymentPolicyInvalid
	}
	policy := DeploymentPolicy{
		environmentID: environmentID,
		mode:          mode,
		createdAt:     createdAt,
	}
	switch mode {
	case DeploymentModeOnPremise:
		if maximumCustomerTenants == nil || *maximumCustomerTenants != 0 {
			return DeploymentPolicy{}, ErrDeploymentPolicyInvalid
		}
		policy.hasCustomerLimit = true
	case DeploymentModeSaaS:
		if maximumCustomerTenants != nil {
			return DeploymentPolicy{}, ErrDeploymentPolicyInvalid
		}
	default:
		return DeploymentPolicy{}, ErrDeploymentPolicyInvalid
	}
	return policy, nil
}

func (policy DeploymentPolicy) EnvironmentID() string { return policy.environmentID }
func (policy DeploymentPolicy) Mode() DeploymentMode  { return policy.mode }
func (policy DeploymentPolicy) CreatedAt() time.Time  { return policy.createdAt }

func (policy DeploymentPolicy) MaximumCustomerTenants() (int64, bool) {
	return policy.maximumCustomerTenants, policy.hasCustomerLimit
}

func (policy DeploymentPolicy) IsValid() bool {
	var maximum *int64
	if policy.hasCustomerLimit {
		value := policy.maximumCustomerTenants
		maximum = &value
	}
	_, err := NewDeploymentPolicy(
		policy.environmentID,
		policy.mode,
		maximum,
		policy.createdAt,
	)
	return err == nil
}

func canonicalUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' ||
		value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
