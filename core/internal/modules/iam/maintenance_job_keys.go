package iam

import maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"

func MaintenanceJobKeys() []MaintenanceJobKey {
	return []MaintenanceJobKey{
		{Kind: string(maintenanceapp.KindCredentialBcryptRetire), PayloadRevision: 1},
		{Kind: string(maintenanceapp.KindDirectoryChecksum), PayloadRevision: 1},
		{Kind: string(maintenanceapp.KindDirectoryRepair), PayloadRevision: 1},
		{Kind: string(maintenanceapp.KindDirectoryTombstonePurge), PayloadRevision: 1},
		{Kind: string(maintenanceapp.KindIdentifierRetire), PayloadRevision: 1},
		{Kind: string(maintenanceapp.KindProfileBackfill), PayloadRevision: 1},
		{Kind: string(maintenanceapp.KindSecurityArtifactPurge), PayloadRevision: 1},
	}
}
