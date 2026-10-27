package postgresv1

import "fmt"

var directDatabaseConnectGrantees = []string{
	"harpy_cluster_admin",
	"harpy_extension_admin",
	"harpy_migrator",
	"harpy_runtime",
	"harpy_v2_runtime",
	"harpy_backup",
	"harpy_restore",
	"harpy_reactivation_observer",
}

func validateDirectDatabaseConnectGrants(grants []ACLGrant, databaseName string) error {
	found := make(map[string]bool, len(directDatabaseConnectGrantees))
	for _, grant := range grants {
		if grant.Object.Kind != "database" || grant.Object.DatabaseName != databaseName ||
			grant.Object.ObjectName != databaseName || grant.Grantor != "harpy_owner" ||
			grant.GrantOption || len(grant.Privileges) != 1 || grant.Privileges[0] != "CONNECT" {
			continue
		}
		found[grant.Grantee] = true
	}
	for _, grantee := range directDatabaseConnectGrantees {
		if !found[grantee] {
			return fmt.Errorf("database CONNECT grant for %s is required", grantee)
		}
	}
	return nil
}
