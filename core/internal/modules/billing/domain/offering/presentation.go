package offering

import (
	"strings"
	"unicode/utf8"
)

const (
	maximumDescriptionBytes = 2048
	maximumHighlightBytes   = 50
	maximumColorBytes       = 20
)

type BillingInterval string

const (
	MonthlyInterval BillingInterval = "monthly"
	AnnualInterval  BillingInterval = "annual"
)

type Presentation struct {
	description     string
	sortOrder       int32
	highlightLabel  string
	color           string
	billingInterval BillingInterval
}

func NewPresentation(
	description string,
	sortOrder int32,
	highlightLabel string,
	color string,
	billingInterval BillingInterval,
) (Presentation, error) {
	value := Presentation{description: description, sortOrder: sortOrder,
		highlightLabel: highlightLabel, color: color, billingInterval: billingInterval}
	if !value.Valid() {
		return Presentation{}, ErrVersion
	}
	return value, nil
}

func (value Presentation) Description() string              { return value.description }
func (value Presentation) SortOrder() int32                 { return value.sortOrder }
func (value Presentation) HighlightLabel() string           { return value.highlightLabel }
func (value Presentation) Color() string                    { return value.color }
func (value Presentation) BillingInterval() BillingInterval { return value.billingInterval }

func (value Presentation) Valid() bool {
	return validOptionalText(value.description, maximumDescriptionBytes, false) &&
		validOptionalText(value.highlightLabel, maximumHighlightBytes, true) &&
		validOptionalText(value.color, maximumColorBytes, true) &&
		(value.billingInterval == MonthlyInterval || value.billingInterval == AnnualInterval)
}

func validOptionalText(value string, maximum int, trim bool) bool {
	if !utf8.ValidString(value) || len(value) > maximum {
		return false
	}
	return !trim || strings.TrimSpace(value) == value
}
