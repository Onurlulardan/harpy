package recordkernel

import "errors"

var (
	ErrDefinitionInvalid = errors.New("Forms record definition is invalid")
	ErrCatalogInvalid    = errors.New("Forms record catalog is invalid")
	ErrInputInvalid      = errors.New("Forms record input is invalid")
	ErrFieldForbidden    = errors.New("Forms record field is forbidden")
	ErrValidation        = errors.New("Forms record validation failed")
	ErrCalculation       = errors.New("Forms record calculation failed")
	ErrBudgetExceeded    = errors.New("Forms record budget exceeded")
)
