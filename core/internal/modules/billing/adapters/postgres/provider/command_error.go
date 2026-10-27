package provider

import (
	"errors"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

type commandStateError interface{ SQLState() string }

func classifyCommandError(err error) error {
	var databaseError commandStateError
	if !errors.As(err, &databaseError) {
		return err
	}
	switch databaseError.SQLState() {
	case "22023":
		return providerapp.ErrCommandRequest
	case "P5731":
		return providerapp.ErrCommandTarget
	case "23505", "P5730":
		return providerapp.ErrCommandConflict
	case "P5721":
		return providerapp.ErrCommandState
	default:
		return err
	}
}
