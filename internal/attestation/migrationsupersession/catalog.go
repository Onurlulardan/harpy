package migrationsupersession

const (
	CatalogFileName        = "migration-supersessions.json"
	ContractVersion        = ContractVersionV1
	ContractVersionV1      = "migration-supersession.v1"
	ContractVersionV2      = "migration-supersession.v2"
	ContractVersionV3      = "migration-supersession.v3"
	maximumCatalogBytes    = 64 * 1024
	maximumCatalogEntries  = 64
	maximumSourceDocuments = 4096
)

type Catalog struct {
	contractVersion string
	owner           string
	scope           string
	sourceDirectory string
	entries         []Entry
}

type Entry struct {
	oldPath           string
	oldSHA256         string
	replacementPath   string
	replacementSHA256 string
}

type catalogDocument struct {
	ContractVersion string          `json:"contract_version"`
	Owner           string          `json:"owner"`
	Scope           string          `json:"scope"`
	SourceDirectory string          `json:"source_directory"`
	Entries         []entryDocument `json:"entries"`
}

type entryDocument struct {
	OldPath           string `json:"old_path"`
	OldSHA256         string `json:"old_sha256"`
	ReplacementPath   string `json:"replacement_path"`
	ReplacementSHA256 string `json:"replacement_sha256"`
}

func (catalog Catalog) Owner() string { return catalog.owner }

func (catalog Catalog) Scope() string { return catalog.scope }

func (catalog Catalog) SourceDirectory() string { return catalog.sourceDirectory }

func (catalog Catalog) ContractVersion() string { return catalog.contractVersion }

func (catalog Catalog) Entries() []Entry {
	return append([]Entry(nil), catalog.entries...)
}

func (entry Entry) OldPath() string { return entry.oldPath }

func (entry Entry) OldSHA256() string { return entry.oldSHA256 }

func (entry Entry) ReplacementPath() string { return entry.replacementPath }

func (entry Entry) ReplacementSHA256() string { return entry.replacementSHA256 }
