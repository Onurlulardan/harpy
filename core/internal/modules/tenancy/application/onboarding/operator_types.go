package onboarding

import (
	"context"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type OperatorActivationOwner = ActivationOwner

const (
	OperatorOwnerTenancy           = ActivationOwnerTenancy
	OperatorOwnerIAM               = ActivationOwnerIAM
	OperatorOwnerBilling           = ActivationOwnerBilling
	OperatorOwnerSystemResources   = ActivationOwnerSystemResources
	OperatorOwnerPlatformOps       = ActivationOwnerPlatformOps
	OperatorOwnerPlatformAuthority = ActivationOwnerPlatformAuthority
	OperatorOwnerAsync             = ActivationOwnerAsync
)

type OperatorActivationCommand string

const (
	OperatorActivateTenant            OperatorActivationCommand = "tenancy.activate_tenant"
	OperatorInitializeTenantIdentity  OperatorActivationCommand = "iam.initialize_tenant_identity"
	OperatorAssignTenantBaseOffering  OperatorActivationCommand = "billing.assign_tenant_base_offering"
	OperatorOpenTenantUsage           OperatorActivationCommand = "billing.open_tenant_usage"
	OperatorInitializeSystemBaseline  OperatorActivationCommand = "systemresources.initialize_tenant_baseline"
	OperatorInitializeRecoverySigning OperatorActivationCommand = "platformops.initialize_recovery_signing_identity"
	OperatorBindInstanceOwner         OperatorActivationCommand = "platformauthority.bind_owner"
	OperatorPinAuthorityHead          OperatorActivationCommand = "platformauthority.pin_protected_authority_head"
	OperatorCompleteActivation        OperatorActivationCommand = "activation.complete"
)

type OperatorActivationRow = ActivationOwnerRow
type OperatorActivationExecutor = ActivationOwnerExecutor

type OperatorActivationResources struct {
	BindOwner func(OperatorActivationOwner) (OperatorActivationExecutor, error)
	Consume   func(OperatorActivationCommand) error
}

type OperatorActivationBinding struct {
	OperationID, EnvironmentID, TenantID              tenantapi.UUID
	FirstUserID, RootOrganizationID                   tenantapi.UUID
	ProtectedAdminBindingID, InstanceOwnerBindingID   tenantapi.UUID
	Kind, DeploymentMode, BindingPurpose, SchemaName  string
	RequestSHA256, ManifestSHA256, StepPlanSHA256     string
	MigrationReceiptID                                tenantapi.UUID
	MigrationReceiptSHA256, AuthorityAckSHA256        string
	OfferingVersionID                                 tenantapi.UUID
	OfferingRevision                                  int64
	OfferingSHA256                                    string
	OfferingEffectiveAt                               time.Time
	OfferingSourceSequence                            int64
	OfferingSourceSHA256                              string
	OfferingSourcePresent                             bool
	RecoverySigningGeneration                         int64
	RecoveryTrustBundleSHA256                         string
	RecoverySigningPresent                            bool
	AuthorityRegistryBindingID, AuthorityScopeID      tenantapi.UUID
	AuthorityScopeKind                                string
	AuthorityGeneration                               int64
	AuthorityPreviousHeadSHA256, AuthorityStateSHA256 string
	AuthorityProposalSHA256                           string
}

type OperatorActivationCallback func(
	context.Context,
	OperatorActivationBinding,
	OperatorActivationResources,
) error
