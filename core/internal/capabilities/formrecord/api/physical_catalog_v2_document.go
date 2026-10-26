package recordkernel

const PhysicalCatalogV2Contract = "forms-record-physical-catalog.v2"

type physicalHandleDocumentV2 struct {
	LogicalID                   string  `json:"logical_id"`
	FieldID                     *string `json:"field_id"`
	Kind                        string  `json:"kind"`
	PhysicalName                string  `json:"physical_name"`
	StorageType                 string  `json:"storage_type"`
	Variant                     string  `json:"variant"`
	Generation                  uint32  `json:"generation"`
	Profile                     string  `json:"profile"`
	TargetBindingKind           string  `json:"target_binding_kind,omitempty"`
	TargetPublicationID         *string `json:"target_publication_id,omitempty"`
	TargetFormID                *string `json:"target_form_id,omitempty"`
	TargetFormVersionID         *string `json:"target_form_version_id,omitempty"`
	TargetRecordCatalogSHA256   *string `json:"target_record_catalog_sha256,omitempty"`
	TargetQueryCatalogSHA256    *string `json:"target_query_catalog_sha256,omitempty"`
	TargetPhysicalCatalogSHA256 *string `json:"target_physical_catalog_sha256,omitempty"`
	TargetPhysicalName          *string `json:"target_physical_name,omitempty"`
}

type physicalCatalogDocumentV2 struct {
	Contract             string                     `json:"contract"`
	PublicationID        string                     `json:"publication_id"`
	PublicationRevision  int64                      `json:"publication_revision"`
	FormID               string                     `json:"form_id"`
	FormVersionID        string                     `json:"form_version_id"`
	RecordCatalogSHA256  string                     `json:"record_catalog_sha256"`
	MappingCatalogSHA256 string                     `json:"mapping_catalog_sha256"`
	Handles              []physicalHandleDocumentV2 `json:"handles"`
}
