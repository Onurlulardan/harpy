package api

import "context"

type PrimaryTenantTargetQuery interface {
	GetPrimaryTenantTarget(context.Context) (PrimaryTenantTarget, error)
}
