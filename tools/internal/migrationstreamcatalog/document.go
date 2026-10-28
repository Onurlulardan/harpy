package migrationstreamcatalog

const (
	rootDocumentPath     = "contracts/storage/migration-streams.yaml"
	streamDirectoryPath  = "contracts/storage/migration-streams"
	catalogVersion       = 2
	ownerDocumentVersion = 1
	maximumIncludes      = 256
	maximumRootBytes     = 64 * 1024
	maximumFragmentBytes = 1024 * 1024
	maximumCatalogBytes  = 16 * 1024 * 1024
)

type rootDocument struct {
	Version  int      `json:"version"`
	Includes []string `json:"includes"`
}

type ownerDocument struct {
	Version int      `json:"version"`
	Owner   string   `json:"owner"`
	Streams []Stream `json:"streams"`
}
