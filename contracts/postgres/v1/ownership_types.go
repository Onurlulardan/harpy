package postgresv1

type OwnershipManifest struct {
	ContractVersion   string                     `json:"contract_version"`
	ManifestID        string                     `json:"manifest_id"`
	DatabaseOrigin    string                     `json:"database_origin"`
	DatabaseName      string                     `json:"database_name"`
	OwnershipExecutor string                     `json:"ownership_executor"`
	Format            string                     `json:"format"`
	ObjectCount       int64                      `json:"object_count"`
	Objects           []OwnershipObject          `json:"objects,omitempty" harpyOptional:"allowed"`
	Chunks            []OwnershipChunkDescriptor `json:"chunks,omitempty" harpyOptional:"allowed"`
}

type OwnershipChunkDescriptor struct {
	Ordinal     int    `json:"ordinal"`
	Count       int    `json:"count"`
	LengthBytes int64  `json:"length_bytes"`
	SHA256      string `json:"sha256"`
	Path        string `json:"path"`
}

type OwnershipChunk struct {
	ContractVersion string            `json:"contract_version"`
	ManifestID      string            `json:"manifest_id"`
	DatabaseOrigin  string            `json:"database_origin"`
	DatabaseName    string            `json:"database_name"`
	Ordinal         int               `json:"ordinal"`
	Count           int               `json:"count"`
	ObjectCount     int64             `json:"object_count"`
	Objects         []OwnershipObject `json:"objects"`
}

type OwnershipObject struct {
	Object         CatalogObject `json:"object"`
	Owner          string        `json:"owner"`
	Classification string        `json:"classification"`
	Executor       string        `json:"executor"`
}
