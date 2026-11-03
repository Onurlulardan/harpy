//go:build integration

package postgres

var iamPortalSetupReadMigrationNames = []string{
	"20260730145000_create_iam_portal_setup_read_v1",
	"20260730145001_grant_iam_portal_setup_read_v1",
}

var iamProfileCommandMigrationNames = []string{
	"20260716013800_create_iam_profile_actor_assert_v1",
	"20260716013801_create_iam_profile_retirement_evidence",
	"20260716013802_create_iam_profile_schema_apply_v1",
	"20260716013803_create_iam_profile_schema_publish_v1",
	"20260716013804_create_iam_profile_document_backfill_v1",
	"20260716013805_create_iam_profile_schema_retire_v1",
	"20260716013806_grant_iam_profile_commands_v1",
}

var iamSelfProfileMigrationNames = []string{
	"20260716015910_create_iam_self_profile_read_v1",
	"20260716015911_create_iam_self_profile_update_v1",
	"20260716015912_grant_iam_self_profile_runtime_v1",
}

var iamRBACHTTPMigrationNames = []string{
	"20260716015700_create_iam_rbac_assignment_groups",
	"20260716015701_create_iam_rbac_assignment_members",
	"20260716015702_create_iam_rbac_assignment_helpers_v1",
	"20260716015703_create_iam_rbac_assignment_binding_v1",
	"20260716015704_create_iam_rbac_assignment_create_v1",
	"20260716015705_create_iam_rbac_assignment_update_v1",
	"20260716015706_create_iam_rbac_assignment_retire_v1",
	"20260716015707_create_iam_rbac_assignment_read_v1",
	"20260716015708_create_iam_rbac_effective_read_v1",
	"20260716015709_grant_iam_rbac_assignment_runtime_v1",
}

var iamRestQueryMigrationNames = []string{
	"20260716014100_create_iam_rest_query_actor_versions_v1",
	"20260716014101_create_iam_rest_query_users_v1",
	"20260716014102_create_iam_rest_query_organizations_v1",
	"20260716014103_create_iam_rest_query_roles_v1",
	"20260716014104_create_iam_rest_query_profile_indexes",
	"20260716014105_grant_iam_rest_query_runtime_v1",
	"20260730113000_create_iam_rest_query_search_views_v2",
	"20260730120010_create_iam_organization_owner_query_v3",
}
