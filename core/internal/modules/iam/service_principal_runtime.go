package iam

import (
	servicepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/serviceprincipal"
	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type ServicePrincipalCapabilityDefinition = serviceapp.CapabilityDefinition
type ServicePrincipalCreateCommand = serviceapp.CreateCommand
type ServicePrincipalStateCommand = serviceapp.StateCommand
type ServicePrincipalGrantCommand = serviceapp.GrantCommand
type ServicePrincipalRevokeCommand = serviceapp.RevokeCommand
type ServicePrincipalPinnedRequest = serviceapp.PinnedRequest
type ServicePrincipalVerificationEvidence = serviceapp.VerificationEvidence
type ServicePrincipalCapabilityKey = serviceapp.CapabilityKey
type ServicePrincipalResourceScope = serviceapp.ResourceScope
type ServicePrincipalScopeKind = serviceapp.ScopeKind
type ServicePrincipalMutationResult = serviceapp.MutationResult

const ServicePrincipalScopeResource = serviceapp.ScopeResource

func ParseServicePrincipalCapabilityKey(raw string) (ServicePrincipalCapabilityKey, error) {
	return serviceapp.ParseCapabilityKey(raw)
}

type ServicePrincipalRuntime struct {
	commands  *serviceapp.Commands
	verifier  serviceapp.Verifier
	authority *serviceapp.PublicationAuthorityApplication
}

func NewServicePrincipalRuntime(
	catalogs platformasync.WriteCatalogs,
	definitions []ServicePrincipalCapabilityDefinition,
	now RBACClock,
) (*ServicePrincipalRuntime, error) {
	catalog, err := serviceapp.NewCapabilityCatalog(definitions)
	if err != nil {
		return nil, err
	}
	store, err := servicepostgres.NewCommandStore()
	if err != nil {
		return nil, err
	}
	rbacSink, err := NewRBACChangeSink(catalogs, now)
	if err != nil {
		return nil, err
	}
	commands, err := serviceapp.NewCommands(
		store, catalog, rbacSink,
	)
	if err != nil {
		return nil, err
	}
	authorityStore, err := servicepostgres.NewPublicationAuthorityStore()
	if err != nil {
		return nil, err
	}
	authority, err := serviceapp.NewPublicationAuthorityApplication(
		commands, authorityStore,
	)
	if err != nil {
		return nil, err
	}
	evidence, err := servicepostgres.NewEvidenceReader()
	if err != nil {
		return nil, err
	}
	verifier, err := serviceapp.NewVerifier(evidence, serviceapp.Clock(now))
	if err != nil {
		return nil, err
	}
	return &ServicePrincipalRuntime{
		commands: commands, verifier: verifier, authority: authority,
	}, nil
}

func (runtime *ServicePrincipalRuntime) Valid() bool {
	return runtime != nil && runtime.commands != nil && runtime.authority != nil
}
