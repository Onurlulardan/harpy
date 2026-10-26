package postgresv1

type ExtensionManifest struct {
	ContractVersion       string              `json:"contract_version"`
	DatabaseOrigin        string              `json:"database_origin"`
	PostgresImageDigest   string              `json:"postgres_image_digest"`
	PostgresEngineVersion string              `json:"postgres_engine_version"`
	ControlPathPolicy     ControlPathPolicy   `json:"control_path_policy"`
	Extensions            []PostgresExtension `json:"extensions"`
}

type ControlPathPolicy struct {
	Resolver           string `json:"resolver"`
	RelativeDirectory  string `json:"relative_directory"`
	AllowAbsolutePaths bool   `json:"allow_absolute_paths"`
	AllowSymlinks      bool   `json:"allow_symlinks"`
}

type PostgresExtension struct {
	Name              string   `json:"name"`
	Version           string   `json:"version"`
	Install           bool     `json:"install"`
	Schema            string   `json:"schema"`
	Owner             string   `json:"owner"`
	ControlFile       string   `json:"control_file"`
	ControlFileSHA256 string   `json:"control_file_sha256"`
	Capabilities      []string `json:"capabilities"`
}
