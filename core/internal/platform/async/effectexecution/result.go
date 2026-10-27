package effectexecution

import (
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
)

type certainty uint8

const (
	certaintyInvalid certainty = iota
	certaintyDefinitive
	certaintyUnknown
)

type Result struct {
	certainty certainty
	receipt   effectsv1.Receipt
}

func Definitive(receipt effectsv1.Receipt) (Result, error) {
	if err := receipt.Validate(); err != nil {
		return Result{}, fmt.Errorf("effect provider receipt: %w", err)
	}
	return Result{certainty: certaintyDefinitive, receipt: receipt}, nil
}

func Unknown() Result { return Result{certainty: certaintyUnknown} }

func (result Result) IsUnknown() bool { return result.certainty == certaintyUnknown }

func (result Result) Receipt() (effectsv1.Receipt, bool) {
	if result.certainty != certaintyDefinitive {
		return effectsv1.Receipt{}, false
	}
	return cloneReceipt(result.receipt), true
}

func (result Result) validateFor(request effectsv1.Request) error {
	if result.certainty == certaintyUnknown {
		return nil
	}
	if result.certainty != certaintyDefinitive {
		return fmt.Errorf("effect provider result certainty is invalid")
	}
	return result.receipt.ValidateForRequest(request)
}

func cloneReceipt(receipt effectsv1.Receipt) effectsv1.Receipt {
	document, err := receipt.MarshalJSON()
	if err != nil {
		return effectsv1.Receipt{}
	}
	copy, err := effectsv1.DecodeReceipt(document)
	if err != nil {
		return effectsv1.Receipt{}
	}
	return copy
}
