package authorization

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
)

type VisibilityMode uint8

const (
	VisibilityAll VisibilityMode = iota + 1
	VisibilityNone
	VisibilityPredicate
)

type Visibility struct {
	resourceRequirement [sha256.Size]byte
	mode                VisibilityMode
	filter              *MandatoryFilter
	validated           bool
}

func NewAllVisibility(resourceRequirement [sha256.Size]byte) (Visibility, error) {
	return newVisibility(resourceRequirement, VisibilityAll, nil)
}

func NewNoVisibility(resourceRequirement [sha256.Size]byte) (Visibility, error) {
	return newVisibility(resourceRequirement, VisibilityNone, nil)
}

func NewPredicateVisibility(
	resourceRequirement [sha256.Size]byte,
	filter MandatoryFilter,
) (Visibility, error) {
	return newVisibility(resourceRequirement, VisibilityPredicate, &filter)
}

func newVisibility(
	resourceRequirement [sha256.Size]byte,
	mode VisibilityMode,
	filter *MandatoryFilter,
) (Visibility, error) {
	visibility := Visibility{
		resourceRequirement: resourceRequirement, mode: mode, validated: true,
	}
	if filter != nil {
		owned := filter.clone()
		visibility.filter = &owned
	}
	if err := visibility.Validate(); err != nil {
		return Visibility{}, err
	}
	return visibility, nil
}

func (visibility Visibility) Validate() error {
	if !visibility.validated || visibility.resourceRequirement == [sha256.Size]byte{} ||
		visibility.mode.Validate() != nil {
		return fmt.Errorf("REST Query visibility is invalid")
	}
	if visibility.mode == VisibilityPredicate {
		if visibility.filter == nil || visibility.filter.Validate() != nil {
			return fmt.Errorf("REST Query visibility is invalid")
		}
	} else if visibility.filter != nil {
		return fmt.Errorf("REST Query visibility is invalid")
	}
	return nil
}

func (mode VisibilityMode) String() string {
	switch mode {
	case VisibilityAll:
		return "all"
	case VisibilityNone:
		return "none"
	case VisibilityPredicate:
		return "predicate"
	default:
		return ""
	}
}

func (mode VisibilityMode) Validate() error {
	if mode.String() == "" {
		return fmt.Errorf("REST Query visibility mode is invalid")
	}
	return nil
}

func (Visibility) String() string   { return authorizationRedaction }
func (Visibility) GoString() string { return authorizationRedaction }
func (Visibility) LogValue() slog.Value {
	return authorizationLogValue()
}
func (visibility Visibility) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (Visibility) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }
func (VisibilityMode) MarshalJSON() ([]byte, error) {
	return rejectAuthorizationJSON()
}
