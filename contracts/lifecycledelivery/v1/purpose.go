package lifecycledeliveryv1

import "errors"

var ErrPurpose = errors.New("IAM lifecycle delivery purpose is invalid")

type Purpose string

const (
	PurposeEmailVerification Purpose = "email_verification_v1"
	PurposePasswordReset     Purpose = "password_reset_v1"
	PurposePortalSetup       Purpose = "portal_setup_v1"
	PurposeInitialAdminSetup Purpose = "initial_admin_setup.v1"
)

func (purpose Purpose) Validate() error {
	switch purpose {
	case PurposeEmailVerification, PurposePasswordReset, PurposePortalSetup,
		PurposeInitialAdminSetup:
		return nil
	default:
		return ErrPurpose
	}
}
