//go:build integration

package migrator

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func builtEntrypointOwnershipManifest() postgresv1.OwnershipManifest {
	object := postgresv1.CatalogObject{
		Kind: "database", DatabaseName: builtEntrypointDatabaseName,
		ObjectName: builtEntrypointDatabaseName,
	}
	return postgresv1.OwnershipManifest{
		ContractVersion: "postgres-ownership-manifest.v1",
		ManifestID:      "33333333-3333-4333-8333-333333333333",
		DatabaseOrigin:  "fresh", DatabaseName: builtEntrypointDatabaseName,
		OwnershipExecutor: "cluster_bootstrap", Format: "inline", ObjectCount: 1,
		Objects: []postgresv1.OwnershipObject{{
			Object: object, Owner: "harpy_owner",
			Classification: "first_party", Executor: "cluster_bootstrap",
		}},
	}
}
