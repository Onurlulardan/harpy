package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type contractRetirementSQLState interface{ SQLState() string }

func classifyContractRetirementError(err error) error {
	if err == nil {
		return nil
	}
	var state contractRetirementSQLState
	if !errors.As(err, &state) {
		return contractretirement.NewFault(contractretirement.FaultDependency)
	}
	switch state.SQLState() {
	case "P5501":
		return contractretirement.NewFault(contractretirement.FaultMissing)
	case "P5502":
		return contractretirement.NewFault(contractretirement.FaultIdentityMismatch)
	case "P5503":
		return contractretirement.NewFault(contractretirement.FaultRetired)
	case "P5504", "22003":
		return contractretirement.NewFault(contractretirement.FaultOverflow)
	case "P5505":
		return contractretirement.NewFault(contractretirement.FaultRevisionMismatch)
	case "P5506":
		return contractretirement.NewFault(contractretirement.FaultNotEligible)
	default:
		return contractretirement.NewFault(contractretirement.FaultDependency)
	}
}
