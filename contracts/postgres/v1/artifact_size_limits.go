package postgresv1

const (
	MaximumMigrationManifestBytes          = 16 * 1024 * 1024
	MaximumExtensionManifestBytes          = 64 * 1024
	MaximumRolesManifestBytes              = 64 * 1024
	MaximumPrivilegesManifestBytes         = 64 * 1024 * 1024
	MaximumOwnershipManifestBytes          = 64 * 1024 * 1024
	MaximumOwnershipChunkBytes             = 4 * 1024 * 1024
	MaximumBootstrapOperationBytes         = 64 * 1024
	MaximumExtensionOperationBytes         = 64 * 1024
	MaximumCredentialCleanupOperationBytes = 64 * 1024
)
