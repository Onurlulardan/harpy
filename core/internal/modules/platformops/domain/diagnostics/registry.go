package diagnostics

var registry = map[ID]Definition{
	DatabaseStats: {
		ID:                 DatabaseStats,
		OutputSchemaSHA256: "b8ec3186878fb0ca71fcd87c52cba83dfa16f2744be9f29cce29e6f026d0b93c",
		MaximumRows:        16, MaximumBytes: 16 * 1024,
	},
	SchemaMetadata: {
		ID:                 SchemaMetadata,
		OutputSchemaSHA256: "8b1b648c0bdcb88228d2ddb7f4699b090e57f55831b75f80c42a7a99fde6ea1a",
		MaximumRows:        256, MaximumBytes: 128 * 1024,
	},
	RecoveryReadiness: {
		ID:                 RecoveryReadiness,
		OutputSchemaSHA256: "c983c74a44433bd0590d406437e35132426e11478cbd45713866392f08e4137e",
		MaximumRows:        32, MaximumBytes: 32 * 1024,
	},
}
