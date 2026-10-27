//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMOrganizationRoleConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const schema = "t_iam_a"
	const (
		organizationOne = "71000000-0000-4000-8000-000000000001"
		organizationTwo = "71000000-0000-4000-8000-000000000002"
		membershipOne   = "72000000-0000-4000-8000-000000000001"
		membershipTwo   = "72000000-0000-4000-8000-000000000002"
		roleOne         = "73000000-0000-4000-8000-000000000001"
	)

	selfErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMOrganization,
		"71000000-0000-4000-8000-000000000003",
		"Self Organization", "self-organization",
		"71000000-0000-4000-8000-000000000003",
	)
	assertIAMConstraint(t, selfErr, "23514", "iam_organizations_parent_check")

	parentErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMOrganization,
		"71000000-0000-4000-8000-000000000004",
		"Unknown Parent", "unknown-parent",
		"71000000-0000-4000-8000-000000000099",
	)
	assertIAMConstraint(t, parentErr, "23503", "iam_organizations_parent_fk")

	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMOrganization,
		organizationOne, "Organization One", "organization-one", nil,
	)
	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMOrganization,
		organizationTwo, "Organization Two", "organization-two", organizationOne,
	)
	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMMembership,
		membershipOne, organizationOne, iamUserOneID,
	)
	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMMembership,
		membershipTwo, organizationTwo, iamUserOneID,
	)
	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMOrganizationRole,
		roleOne, organizationOne, "Organization One Operator",
	)

	crossOrganizationErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMOrganizationRoleAssignment,
		"74000000-0000-4000-8000-000000000001",
		iamUserOneID, roleOne, organizationOne, membershipTwo,
	)
	assertIAMConstraint(
		t,
		crossOrganizationErr,
		"23503",
		"iam_user_role_assignments_membership_fk",
	)

	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMOrganizationRoleAssignment,
		"74000000-0000-4000-8000-000000000002",
		iamUserOneID, roleOne, organizationOne, membershipOne,
	)
	assertIAMDefaultRoleUniqueness(t, ctx, database, schema, organizationOne)
}
