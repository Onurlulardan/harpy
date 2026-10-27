package onboarding

import onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"

var activationOwners = map[onboardingapp.ActivationOwner]bool{
	onboardingapp.ActivationOwnerTenancy:           true,
	onboardingapp.ActivationOwnerIAM:               true,
	onboardingapp.ActivationOwnerBilling:           true,
	onboardingapp.ActivationOwnerSystemResources:   true,
	onboardingapp.ActivationOwnerPlatformOps:       true,
	onboardingapp.ActivationOwnerPlatformAuthority: true,
	onboardingapp.ActivationOwnerAsync:             true,
}

func (transaction *activationTransaction) bindOwner(
	owner onboardingapp.ActivationOwner,
) (onboardingapp.ActivationOwnerExecutor, error) {
	transaction.mutex.Lock()
	prepared := transaction.schema != ""
	transaction.mutex.Unlock()
	if !transaction.usable() || !prepared || !activationOwners[owner] {
		return onboardingapp.ActivationOwnerExecutor{}, errUnavailable
	}
	return transaction.bind(owner)
}
