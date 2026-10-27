package api

import (
	"context"
	"regexp"
)

var recoverySchemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type RecoveryTarget struct {
	environmentID   string
	tenantID        string
	slug            string
	schemaName      string
	tenantSetSHA256 [32]byte
}

func NewRecoveryTarget(
	environmentID string,
	tenantID string,
	slug string,
	schemaName string,
	tenantSetSHA256 [32]byte,
) (RecoveryTarget, error) {
	if !canonicalUUID(environmentID) || !canonicalUUID(tenantID) ||
		!validCanonicalTenantSlug(slug) || !recoverySchemaPattern.MatchString(schemaName) ||
		tenantSetSHA256 == [32]byte{} {
		return RecoveryTarget{}, ErrPrimaryTenantTargetInvalid
	}
	return RecoveryTarget{environmentID: environmentID, tenantID: tenantID,
		slug: slug, schemaName: schemaName, tenantSetSHA256: tenantSetSHA256}, nil
}

func (target RecoveryTarget) EnvironmentID() string { return target.environmentID }
func (target RecoveryTarget) TenantID() string      { return target.tenantID }
func (target RecoveryTarget) Slug() string          { return target.slug }
func (target RecoveryTarget) SchemaName() string    { return target.schemaName }
func (target RecoveryTarget) TenantSetSHA256() [32]byte {
	return target.tenantSetSHA256
}

type RecoveryTargets interface {
	Get(context.Context) (RecoveryTarget, error)
}
