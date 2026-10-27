package adoption

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var fixtureIDs = struct {
	tenant, org, role, user, allResource, usersResource uuid.UUID
	allAction, readAction, permission, request, group   uuid.UUID
	invitation, schema, field, layout                   uuid.UUID
}{
	uuid.MustParse("10000000-0000-0000-0000-000000000001"),
	uuid.MustParse("10000000-0000-0000-0000-000000000002"),
	uuid.MustParse("10000000-0000-0000-0000-000000000003"),
	uuid.MustParse("10000000-0000-0000-0000-000000000004"),
	uuid.MustParse("10000000-0000-0000-0000-000000000005"),
	uuid.MustParse("10000000-0000-0000-0000-000000000006"),
	uuid.MustParse("10000000-0000-0000-0000-000000000007"),
	uuid.MustParse("10000000-0000-0000-0000-000000000008"),
	uuid.MustParse("10000000-0000-0000-0000-000000000009"),
	uuid.MustParse("10000000-0000-0000-0000-00000000000a"),
	uuid.MustParse("10000000-0000-0000-0000-00000000000b"),
	uuid.MustParse("10000000-0000-0000-0000-00000000000c"),
	uuid.MustParse("10000000-0000-0000-0000-00000000000d"),
	uuid.MustParse("10000000-0000-0000-0000-00000000000e"),
	uuid.MustParse("10000000-0000-0000-0000-00000000000f"),
}

func validFixture() (Snapshot, Mapping) {
	id := fixtureIDs
	snapshot := Snapshot{
		TenantID: id.tenant,
		Organizations: []SourceOrganization{{
			ID: id.org, TenantID: id.tenant, Name: "Forge", Slug: "forge", Status: "ACTIVE",
		}},
		Roles: []SourceRole{{
			ID: id.role, TenantID: id.tenant, Name: "Administrator",
			OrganizationID: id.org.String(), Default: true,
		}},
		Users: []SourceUser{{
			ID: id.user, TenantID: id.tenant, Email: "Admin@Example.com",
			PasswordHash: "$2a$12$" + strings.Repeat("a", 53), RoleIDs: []string{id.role.String()},
			OrganizationID: id.org.String(), Status: "ACTIVE", TokenVersion: 1,
			AccountType: "operator",
		}},
		Resources: []SourceResource{
			{ID: id.allResource, TenantID: id.tenant, Key: "ALL", System: true},
			{ID: id.usersResource, TenantID: id.tenant, Key: "users", System: true},
		},
		Actions: []SourceAction{
			{ID: id.allAction, TenantID: id.tenant, Key: "ALL", System: true},
			{ID: id.readAction, TenantID: id.tenant, Key: "read", System: true},
		},
		Permissions: []SourcePermission{{
			ID: id.permission, TenantID: id.tenant, ResourceID: id.allResource,
			Target: "ROLE", RoleID: pointer(id.role), ActionIDs: []uuid.UUID{id.allAction},
		}},
		PermissionRequests: []SourcePermissionRequest{{
			ID: id.request, RequesterID: id.user, TargetID: id.user,
			ResourceID: id.usersResource, ActionID: id.readAction, State: "PENDING",
			GroupID: pointer(id.group),
		}},
		ProfileFields: []SourceProfileField{{ResourceKind: "user", Key: "nickname", ValueKind: "string"}},
		LayoutFields:  []SourceLayoutField{{ResourceKind: "user", FieldKey: "nickname", SectionKey: "main"}},
		Invitations:   []SourceInvitation{{ID: id.invitation, Email: "next@example.com", RoleID: id.role.String(), State: "PENDING"}},
	}
	return snapshot, validMapping()
}

func pointer(value uuid.UUID) *uuid.UUID { return &value }

func testNormalizer(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || !strings.Contains(value, "@") {
		return "", fmt.Errorf("identifier invalid")
	}
	return value, nil
}
