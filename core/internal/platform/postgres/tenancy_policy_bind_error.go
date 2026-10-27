package postgres

import "errors"

var (
	ErrTenancyPolicyProfile = errors.New(
		"postgres tenancy policy reader requires API or worker profile",
	)
	ErrTenancyPolicyBindingInvalid = errors.New(
		"postgres tenancy policy reader binding is invalid",
	)
)
