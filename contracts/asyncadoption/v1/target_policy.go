package asyncadoptionv1

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func ValidateTargetPolicy(item Item, catalog jobsv1.Catalog) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("validate adoption item: %w", err)
	}
	policy, err := catalog.Resolve(item.TargetKind, item.TargetPayloadRevision)
	if err != nil {
		return fmt.Errorf("resolve adoption target policy: %w", err)
	}
	if string(policy.EffectClass) != string(item.TargetEffectClass) ||
		policy.MaximumAttempts != item.TargetMaximumAttempts {
		return fmt.Errorf("adoption target policy differs from captured item")
	}
	return nil
}
