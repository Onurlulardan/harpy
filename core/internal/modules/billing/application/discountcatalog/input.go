package discountcatalog

import (
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	moneydomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	"github.com/google/uuid"
)

type MoneyInput struct {
	MinorUnits      int64
	Currency        string
	RegistryVersion int64
}

type PolicyInput struct {
	Name               string
	Description        string
	Code               string
	Kind               string
	BasisPoints        *int64
	FixedAmount        *MoneyInput
	Duration           string
	DurationCycles     *int
	Priority           int32
	AllOfferings       bool
	OfferingVersionIDs []string
	ValidStart         time.Time
	ValidEnd           *time.Time
	GlobalUseLimit     *int64
	PerTenantUseLimit  *int64
}

func BuildPresentation(input PolicyInput) (discountdomain.Presentation, error) {
	presentation, err := discountdomain.NewPresentation(input.Name, input.Description)
	if err != nil {
		return discountdomain.Presentation{}, ErrRequest
	}
	return presentation, nil
}

func BuildPolicy(input PolicyInput) (discountdomain.Policy, error) {
	code, codeErr := discountdomain.NormalizeCode(input.Code)
	duration, durationErr := buildInputDuration(input.Duration, input.DurationCycles)
	offeringIDs, offeringErr := buildInputOfferingIDs(input.OfferingVersionIDs)
	if codeErr != nil || durationErr != nil || offeringErr != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	switch input.Kind {
	case string(discountdomain.BasisPoint):
		return buildBasisPointPolicy(input, code, duration, offeringIDs)
	case string(discountdomain.Fixed):
		return buildFixedPolicy(input, code, duration, offeringIDs)
	default:
		return discountdomain.Policy{}, ErrRequest
	}
}

func buildBasisPointPolicy(
	input PolicyInput, code discountdomain.Code, duration discountdomain.Duration,
	offeringIDs []discountdomain.Identifier,
) (discountdomain.Policy, error) {
	if input.BasisPoints == nil || input.FixedAmount != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	rate, err := moneydomain.NewBasisPoints(*input.BasisPoints)
	if err != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	policy, err := discountdomain.NewBasisPointPolicy(
		code, rate, duration, input.Priority, input.GlobalUseLimit,
		input.PerTenantUseLimit, input.ValidStart, input.ValidEnd,
		input.AllOfferings, offeringIDs,
	)
	if err != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	return policy, nil
}

func buildFixedPolicy(
	input PolicyInput, code discountdomain.Code, duration discountdomain.Duration,
	offeringIDs []discountdomain.Identifier,
) (discountdomain.Policy, error) {
	if input.FixedAmount == nil || input.BasisPoints != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	amount, err := moneydomain.New(
		input.FixedAmount.MinorUnits, input.FixedAmount.Currency,
		input.FixedAmount.RegistryVersion,
	)
	if err != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	policy, err := discountdomain.NewFixedPolicy(
		code, amount, duration, input.Priority, input.GlobalUseLimit,
		input.PerTenantUseLimit, input.ValidStart, input.ValidEnd,
		input.AllOfferings, offeringIDs,
	)
	if err != nil {
		return discountdomain.Policy{}, ErrRequest
	}
	return policy, nil
}

func buildInputDuration(value string, cycles *int) (discountdomain.Duration, error) {
	switch value {
	case string(discountdomain.Once):
		if cycles == nil {
			return discountdomain.OnceDuration(), nil
		}
	case string(discountdomain.Forever):
		if cycles == nil {
			return discountdomain.ForeverDuration(), nil
		}
	case string(discountdomain.Repeating):
		if cycles != nil && *cycles > 0 {
			duration, err := discountdomain.RepeatingDuration(int64(*cycles))
			if err == nil {
				return duration, nil
			}
		}
	}
	return discountdomain.Duration{}, ErrRequest
}

func buildInputOfferingIDs(values []string) ([]discountdomain.Identifier, error) {
	result := make([]discountdomain.Identifier, len(values))
	for index, value := range values {
		identifier, err := uuid.Parse(value)
		if err != nil || identifier == uuid.Nil || identifier.Version() != uuid.Version(7) {
			return nil, ErrRequest
		}
		result[index] = discountdomain.Identifier(identifier)
	}
	return result, nil
}
