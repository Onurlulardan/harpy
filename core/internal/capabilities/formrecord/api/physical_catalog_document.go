package recordkernel

const (
	PhysicalCatalogContract     = "forms-record-physical-catalog.v1"
	maximumPhysicalCatalogBytes = 1 << 20
	physicalFormTable           = "form_table"
	physicalFieldColumn         = "field_column"
	physicalRelationTable       = "relation_table"
)

type physicalHandleDocument struct {
	LogicalID    string  `json:"logical_id"`
	FieldID      *string `json:"field_id"`
	Kind         string  `json:"kind"`
	PhysicalName string  `json:"physical_name"`
	StorageType  string  `json:"storage_type"`
	Variant      string  `json:"variant"`
	Generation   uint32  `json:"generation"`
	Profile      string  `json:"profile"`
}

type physicalCatalogDocument struct {
	Contract             string                   `json:"contract"`
	PublicationID        string                   `json:"publication_id"`
	PublicationRevision  int64                    `json:"publication_revision"`
	FormID               string                   `json:"form_id"`
	FormVersionID        string                   `json:"form_version_id"`
	RecordCatalogSHA256  string                   `json:"record_catalog_sha256"`
	MappingCatalogSHA256 string                   `json:"mapping_catalog_sha256"`
	Handles              []physicalHandleDocument `json:"handles"`
}
