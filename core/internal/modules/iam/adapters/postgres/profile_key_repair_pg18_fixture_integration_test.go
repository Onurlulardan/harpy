//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

type iamProfileRepairInstallation struct {
	schema, tenant, operation, user, organization, assignment    string
	userSchema, userField, organizationSchema, organizationField string
	roleSchema, roleField, identifier, accessMode, membership    string
	role, permission, rolePermission, displayName, tenantName    string
	email                                                        string
}

type iamProfileRepairFixtures struct {
	legacy, rolling, badShape, badDigest iamProfileRepairInstallation
}

func prepareIAMProfileKeyRepairPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	origin string,
) iamProfileRepairFixtures {
	t.Helper()
	fixtures := iamProfileRepairFixtures{
		legacy:    newIAMProfileRepairInstallation("t_iam_a", iamFoundationTenants[0].id, 1),
		rolling:   newIAMProfileRepairInstallation("t_iam_b", iamFoundationTenants[1].id, 2),
		badShape:  newIAMProfileRepairInstallation("t_profile_bad_shape", profileRepairID(301, 4), 3),
		badDigest: newIAMProfileRepairInstallation("t_profile_bad_digest", profileRepairID(401, 4), 4),
	}
	_, err := database.migrator.ExecContext(ctx, `
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE TABLE platform.tenants (
 id uuid PRIMARY KEY,schema_name name NOT NULL UNIQUE
);
CREATE SCHEMA t_iam_a AUTHORIZATION harpy_owner;
CREATE SCHEMA t_iam_b AUTHORIZATION harpy_owner;
CREATE SCHEMA t_profile_bad_shape AUTHORIZATION harpy_owner;
CREATE SCHEMA t_profile_bad_digest AUTHORIZATION harpy_owner;
INSERT INTO platform.tenants(id,schema_name) VALUES
 ('10000000-0000-4000-8000-000000000001','t_iam_a'),
 ('20000000-0000-4000-8000-000000000002','t_iam_b'),
 ('0000012d-0000-4000-8000-00000000012d','t_profile_bad_shape'),
 ('00000191-0000-4000-8000-000000000191','t_profile_bad_digest');
GRANT USAGE ON SCHEMA platform,t_iam_a,t_iam_b,
 t_profile_bad_shape,t_profile_bad_digest TO harpy_runtime;`)
	if err != nil {
		t.Fatal(err)
	}
	prepareIAMPlatformAuthPrerequisites(t, ctx, database.migrator)
	applyIAMPlatformDocuments(t, ctx, database.migrator,
		loadIAMFoundationDocuments(t, origin, "platform"))
	tenantFoundation := loadIAMFoundationDocuments(t, origin, "tenant")
	all := []iamProfileRepairInstallation{
		fixtures.legacy, fixtures.rolling, fixtures.badShape, fixtures.badDigest,
	}
	for _, fixture := range all {
		applyIAMTenantDocuments(t, ctx, database.migrator,
			fixture.schema, tenantFoundation)
	}
	for _, names := range [][]string{
		iamLifecycleGuardMigrationNames,
		iamRBACMigrationNames,
		iamProfileRevisionMigrationNames,
		{"20260716014300_create_iam_maintenance_schedules"},
	} {
		documents := iamProfileRepairDocuments(t, origin, names)
		for _, fixture := range all {
			applyIAMTenantDocuments(t, ctx, database.migrator,
				fixture.schema, documents)
		}
	}
	prepareIAMProfileRepairActivationStubPG18(t, ctx, database.migrator)
	predecessorNames := []string{
		"20260722090000_create_iam_installation_profile_v1",
		"20260722090001_create_iam_installation_identity_facts",
		"20260722090002_create_iam_installation_identity_core_v1",
		"20260722090003_create_iam_installation_protected_admin_v1",
		"20260722090004_create_iam_installation_initialize_v1",
		"20260722090005_create_iam_installation_schedules_v1",
		"20260729100025_create_iam_installation_tenant_capability_entrypoints_v1",
	}
	documents := iamProfileRepairDocuments(t, origin, predecessorNames)
	for _, fixture := range all {
		applyIAMTenantDocuments(t, ctx, database.migrator,
			fixture.schema, documents)
	}
	for _, fixture := range []iamProfileRepairInstallation{
		fixtures.legacy, fixtures.badShape, fixtures.badDigest,
	} {
		replayed, callErr := callIAMInstallationPG18(
			ctx, database.runtime, fixture, "v2",
		)
		if callErr != nil || replayed {
			t.Fatalf("%s predecessor v2 = %t: %v",
				fixture.schema, replayed, callErr)
		}
	}
	return fixtures
}

func iamProfileRepairDocuments(
	t *testing.T, origin string, names []string,
) []iamFoundationDocument {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(names))
	for _, name := range names {
		documents = append(documents, iamFoundationDocument{
			path: name, content: readIAMIdentityMigration(t, origin, name),
		})
	}
	return documents
}

func prepareIAMProfileRepairActivationStubPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE FUNCTION platform.iam_installation_activation_assert_v2(
 uuid,uuid,uuid,name
) RETURNS boolean LANGUAGE sql SECURITY DEFINER
SET search_path TO pg_catalog AS 'SELECT true';
REVOKE ALL ON FUNCTION platform.iam_installation_activation_assert_v2(
 uuid,uuid,uuid,name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_activation_assert_v2(
 uuid,uuid,uuid,name) TO harpy_runtime;`)
	if err != nil {
		t.Fatal(err)
	}
}

func newIAMProfileRepairInstallation(
	schema, tenant string, seed int,
) iamProfileRepairInstallation {
	next := func(offset, version int) string {
		return profileRepairID(seed*100+offset, version)
	}
	return iamProfileRepairInstallation{
		schema: schema, tenant: tenant, operation: next(1, 4),
		user: next(2, 4), organization: next(3, 4), assignment: next(4, 4),
		userSchema:         iamInstallationDerivedID(next(1, 4), "user-profile-schema"),
		userField:          iamInstallationDerivedID(next(1, 4), "user-display-name-field"),
		organizationSchema: next(7, 7), organizationField: next(8, 7),
		roleSchema: next(9, 7), roleField: next(10, 7),
		identifier: next(11, 4), accessMode: next(12, 4),
		membership: next(13, 4), role: next(14, 4),
		permission: next(15, 4), rolePermission: next(16, 4),
		email:       fmt.Sprintf("owner-%d@harpy.test", seed),
		displayName: "Yerel \"Sahip\"\nİ😀",
		tenantName:  fmt.Sprintf("Tenant %d", seed),
	}
}

func profileRepairID(value, version int) string {
	return fmt.Sprintf("%08x-0000-%d000-8000-%012x", value, version, value)
}

func callIAMInstallationPG18(
	ctx context.Context, database *sql.DB,
	fixture iamProfileRepairInstallation, version string,
) (bool, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx,
		"SET search_path TO "+fixture.schema+",pg_catalog"); err != nil {
		return false, err
	}
	query := `SELECT iam_installation_initialize_` + version + `(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`
	var replayed bool
	err = connection.QueryRowContext(ctx, query,
		fixture.operation, "30000000-0000-4000-8000-000000000001",
		fixture.tenant, fixture.user, fixture.organization, fixture.assignment,
		fixture.email, fixture.displayName, fixture.tenantName,
		strings.Repeat("ab", 32), fixture.userSchema, fixture.userField,
		fixture.organizationSchema, fixture.organizationField,
		fixture.roleSchema, fixture.roleField, fixture.identifier,
		fixture.accessMode, fixture.membership, fixture.role,
		fixture.permission, fixture.rolePermission).Scan(&replayed)
	return replayed, err
}
