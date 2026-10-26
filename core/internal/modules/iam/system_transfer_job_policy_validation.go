package iam

import (
	"errors"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

var ErrSystemTransferJobPolicy = errors.New("IAM system transfer job policy is invalid")

func ValidateSystemExportJobPolicy(policy jobsv1.KindPolicy) error {
	return validateSystemTransferJobPolicy(policy, SystemExportJobKind,
		jobsv1.EffectExternal, 3)
}

func ValidateSystemImportJobPolicy(policy jobsv1.KindPolicy) error {
	return validateSystemTransferJobPolicy(policy, SystemImportJobKind,
		jobsv1.EffectDatabaseTransactional, 8)
}

func ValidateSystemExportReconcileJobPolicy(policy jobsv1.KindPolicy) error {
	return validateSystemTransferJobPolicy(policy, SystemExportReconcileJobKind,
		jobsv1.EffectExternal, 8)
}

func validateSystemTransferJobPolicy(
	policy jobsv1.KindPolicy,
	kind string,
	effect jobsv1.EffectClass,
	maximumAttempts int,
) error {
	if policy.Validate() != nil || policy.Kind != kind || policy.PayloadRevision != 1 ||
		policy.OwnerDomain != "iam" || policy.EffectClass != effect ||
		policy.MaximumAttempts != maximumAttempts {
		return ErrSystemTransferJobPolicy
	}
	return nil
}
