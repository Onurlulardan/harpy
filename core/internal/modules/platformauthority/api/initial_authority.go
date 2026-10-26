package api

import (
	"context"
	"time"
)

type DeploymentMode string
type ScopeKind string
type ExpiryPolicy string

const (
	ModeOnPremise    DeploymentMode = "onpremise"
	ModeSaaS         DeploymentMode = "saas"
	ScopeEnvironment ScopeKind      = "environment"
	ScopeTenant      ScopeKind      = "tenant"
	ExpiryOnline     ExpiryPolicy   = "online"
	ExpiryAirGap     ExpiryPolicy   = "airgap"
)

type State struct {
	Schema                         string         `json:"schema"`
	EnvironmentID                  string         `json:"environment_id"`
	DeploymentMode                 DeploymentMode `json:"deployment_mode"`
	RegistryBindingID              string         `json:"registry_binding_id"`
	ScopeKind                      ScopeKind      `json:"scope_kind"`
	ScopeID                        string         `json:"scope_id"`
	ProposalSignerGeneration       int64          `json:"proposal_signer_generation"`
	ProposalSignerKID              string         `json:"proposal_signer_kid"`
	OwnerBindingID                 *string        `json:"owner_binding_id"`
	OwnerBindingGeneration         *int64         `json:"owner_binding_generation"`
	PrimaryAdminBindingSetSHA256   *string        `json:"primary_protected_admin_binding_set_sha256"`
	RecoveryAnchorSetSHA256        *string        `json:"recovery_anchor_set_sha256"`
	RecoveryTrustBundleSHA256      *string        `json:"recovery_trust_bundle_sha256"`
	TenantID                       *string        `json:"tenant_id"`
	ProtectedAdminBindingSetSHA256 *string        `json:"protected_admin_binding_set_sha256"`
}

type InitialAuthorityRequest struct {
	OperationID string
	State       State
	IssuedAt    time.Time
	Policy      ExpiryPolicy
}

type CompensationRequest struct {
	OperationID          string
	State                State
	CompensatesAckSHA256 string
	IssuedAt             time.Time
	Policy               ExpiryPolicy
}

type PreparedAuthority struct {
	OperationID       string
	RegistryBindingID string
	ScopeKind         string
	ScopeID           string
	Generation        int64
	StateSHA256       string
	ProposalSHA256    string
	ProposalBytes     []byte
	ExpiresAt         time.Time
}

type AuthorityReceipt struct {
	OperationID       string
	RegistryBindingID string
	ScopeKind         string
	ScopeID           string
	Generation        int64
	StateSHA256       string
	ProposalSHA256    string
	AckSHA256         string
	RegistrySequence  int64
	CompletedAt       time.Time
}

type InitialAuthority interface {
	PrepareInitial(context.Context, InitialAuthorityRequest) (PreparedAuthority, error)
	AcknowledgePrepared(context.Context, string) (AuthorityReceipt, error)
	CompensateUncommitted(context.Context, CompensationRequest) (AuthorityReceipt, error)
}
