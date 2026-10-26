package licenseauthority

import (
	"context"
	"crypto/ed25519"
	"time"

	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

type DeliveryKind = licensev2.DeliveryKind

const (
	DeliveryConnected   = licensev2.DeliveryConnected
	DeliveryAirGap      = licensev2.DeliveryAirGap
	DeliveryDevelopment = licensev2.DeliveryDevelopment
)

type Status = licensev2.Status

const (
	StatusActive    = licensev2.StatusActive
	StatusSuspended = licensev2.StatusSuspended
	StatusRevoked   = licensev2.StatusRevoked
	StatusExpired   = licensev2.StatusExpired
)

type RuntimeState = billingapi.LicenseRuntimeState

const (
	RuntimeValid    = billingapi.LicenseRuntimeValid
	RuntimeGrace    = billingapi.LicenseRuntimeGrace
	RuntimeBuildOff = billingapi.LicenseRuntimeBuildOff
)

type Binding struct {
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	EnvironmentID          uuid.UUID
	DeliveryKind           DeliveryKind
}

type KeyResolver interface {
	LicenseKey(context.Context, string, string, time.Time, time.Time) (ed25519.PublicKey, error)
}

type KeyResolverFunc func(
	context.Context, string, string, time.Time, time.Time,
) (ed25519.PublicKey, error)

func (resolve KeyResolverFunc) LicenseKey(
	ctx context.Context, issuer, keyID string, issuedAt, verifiedAt time.Time,
) (ed25519.PublicKey, error) {
	return resolve(ctx, issuer, keyID, issuedAt, verifiedAt)
}

type FeatureClaims struct {
	FormBuilder, WorkflowBuilder, APIBuilder bool
	DashboardBuilder, AIAssistant            bool
}

type LimitClaims struct {
	MaxUsers, MaxForms, MaxWorkflows int64
	MaxAPIEndpoints, MaxDashboards   int64
	MaxTenants                       int64
}
