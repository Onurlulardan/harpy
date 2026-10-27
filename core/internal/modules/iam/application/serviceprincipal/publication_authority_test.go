package serviceprincipal

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPublicationAuthorityCreatesOnceAndReplaysExactGrant(t *testing.T) {
	ctx, scope, actorID, tenantID := publicationAuthorityScope(t)
	port := &publicationAuthorityFixturePort{permission: uuid.New()}
	capability, err := ParseCapabilityKey("workflow.instance.execute")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCapabilityCatalog([]CapabilityDefinition{{
		Key: capability, Owner: "workflow", ResourceKind: "definition",
		ResourceScope: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	commands, err := NewCommands(port, catalog, changeSinkStub{})
	if err != nil {
		t.Fatal(err)
	}
	application, err := NewPublicationAuthorityApplication(commands, port)
	if err != nil {
		t.Fatal(err)
	}
	definitionID := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	purpose, _ := PublicationAuthorityStepUpPurpose(tenantID, definitionID)
	request := ProvisionPublicationAuthority{
		DefinitionID:   definitionID,
		ManifestSHA256: [32]byte{7}, IdempotencyKey: "publication-auth-001",
		MutationEvidence: PublicationAuthorityEvidence{
			ActorID: actorID, ExpectedTokenVersion: 4,
			ExpectedAuthorizationVersion: 5, PurposeSHA256: purpose,
			AuthenticatedAt: time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC),
			AMRPassword:     true,
		},
	}
	first, err := application.Provision(ctx, scope, request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := application.Provision(ctx, scope, request)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || first.ServicePrincipalID.Version() != 7 ||
		first.GrantID.Version() != 7 || first.GrantRevision != 1 ||
		port.createCalls != 1 || port.grantCalls != 1 {
		t.Fatalf("first=%+v second=%+v creates=%d grants=%d",
			first, second, port.createCalls, port.grantCalls)
	}
}

func TestPublicationAuthorityRejectsMutatedReservedGrant(t *testing.T) {
	ctx, scope, actorID, tenantID := publicationAuthorityScope(t)
	port := &publicationAuthorityFixturePort{
		permission: uuid.New(),
		principal: PublicationAuthorityPrincipal{
			ID: publicationPrincipalID, Found: true, Kind: "workflow", State: "active",
		},
	}
	capability, _ := ParseCapabilityKey("workflow.instance.execute")
	catalog, _ := NewCapabilityCatalog([]CapabilityDefinition{{
		Key: capability, Owner: "workflow", ResourceKind: "definition",
		ResourceScope: true,
	}})
	commands, _ := NewCommands(port, catalog, changeSinkStub{})
	application, _ := NewPublicationAuthorityApplication(commands, port)
	definitionID := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	purpose, _ := PublicationAuthorityStepUpPurpose(tenantID, definitionID)
	request := ProvisionPublicationAuthority{
		DefinitionID:   definitionID,
		ManifestSHA256: [32]byte{7}, IdempotencyKey: "publication-auth-002",
		MutationEvidence: PublicationAuthorityEvidence{
			ActorID: actorID, ExpectedTokenVersion: 4,
			ExpectedAuthorizationVersion: 5, PurposeSHA256: purpose,
			AuthenticatedAt: time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC),
			AMRWebAuthn:     true,
		},
	}
	grantID := publicationAuthorityID(
		request.DefinitionID, request.ManifestSHA256, "grant",
	)
	port.grant = PublicationAuthorityGrant{
		PublicationAuthority: PublicationAuthority{
			ServicePrincipalID: publicationPrincipalID,
			GrantID:            grantID, GrantRevision: 2,
		},
		Found: true, State: "active", Capability: capability,
		Scope: ResourceScope{Kind: ScopeResource, Owner: "workflow",
			ResourceKind: "definition", ResourceID: request.DefinitionID},
		ManifestSHA256: request.ManifestSHA256,
	}
	if _, err := application.Provision(ctx, scope, request); !errors.Is(
		err, ErrGrantMismatch,
	) {
		t.Fatalf("mutated reserved grant error=%v", err)
	}
}

func TestPublicationAuthorityRejectsCrossTargetStepUpPurpose(t *testing.T) {
	ctx, scope, actorID, tenantID := publicationAuthorityScope(t)
	port := &publicationAuthorityFixturePort{permission: uuid.New()}
	capability, _ := ParseCapabilityKey("workflow.instance.execute")
	catalog, _ := NewCapabilityCatalog([]CapabilityDefinition{{
		Key: capability, Owner: "workflow", ResourceKind: "definition",
		ResourceScope: true,
	}})
	commands, _ := NewCommands(port, catalog, changeSinkStub{})
	application, _ := NewPublicationAuthorityApplication(commands, port)
	definitionID := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	otherID := uuid.MustParse("018f0000-0000-7000-8000-000000000003")
	wrongPurpose, _ := PublicationAuthorityStepUpPurpose(tenantID, otherID)
	_, err := application.Provision(ctx, scope, ProvisionPublicationAuthority{
		DefinitionID: definitionID, ManifestSHA256: [32]byte{7},
		IdempotencyKey: "publication-auth-003",
		MutationEvidence: PublicationAuthorityEvidence{
			ActorID: actorID, ExpectedTokenVersion: 4,
			ExpectedAuthorizationVersion: 5, PurposeSHA256: wrongPurpose,
			AuthenticatedAt: time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC),
			AMRTOTP:         true,
		},
	})
	if !errors.Is(err, ErrInvalidGrant) || port.createCalls != 0 ||
		port.grantCalls != 0 {
		t.Fatalf("cross-target purpose error=%v creates=%d grants=%d",
			err, port.createCalls, port.grantCalls)
	}
}
