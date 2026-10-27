package profile

import "errors"

var (
	ErrFieldInvalid            = errors.New("IAM profile field is invalid")
	ErrSchemaInvalid           = errors.New("IAM profile schema is invalid")
	ErrValidationInvalid       = errors.New("IAM profile validation is invalid")
	ErrDocumentInvalid         = errors.New("IAM profile document is invalid")
	ErrPatchInvalid            = errors.New("IAM profile patch is invalid")
	ErrRequiredFieldMissing    = errors.New("IAM profile required field is missing")
	ErrFieldUnavailable        = errors.New("IAM profile field is unavailable")
	ErrQueryCapabilityRejected = errors.New("IAM profile query capability is rejected")
	ErrSchemaNotFound          = errors.New("IAM profile schema was not found")
	ErrDocumentNotFound        = errors.New("IAM profile document was not found")
	ErrRecoveryRequired        = errors.New("IAM profile recovery is required")
)
