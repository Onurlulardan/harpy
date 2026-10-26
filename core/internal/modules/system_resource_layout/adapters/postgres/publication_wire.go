package postgres

import "encoding/json"

type publicationWire struct {
	Revision               json.RawMessage `json:"revision"`
	Generation             int64           `json:"generation"`
	SourceDefinitionSHA256 string          `json:"source_definition_sha256"`
	CompilerVersion        string          `json:"compiler_version"`
	CatalogContract        string          `json:"catalog_contract"`
	CatalogJSON            string          `json:"catalog_json"`
	CatalogSHA256          string          `json:"catalog_sha256"`
	PresentationContract   string          `json:"presentation_contract"`
	PresentationJSON       string          `json:"presentation_json"`
	PresentationSHA256     string          `json:"presentation_sha256"`
}
