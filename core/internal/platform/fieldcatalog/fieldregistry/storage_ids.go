package fieldregistry

const (
	StorageLayoutColumn        StorageLayout = "column"
	StorageLayoutRelationTable StorageLayout = "relation_table"
)

const (
	StorageText      StorageType = "text"
	StorageInteger   StorageType = "integer"
	StorageDecimal   StorageType = "decimal"
	StorageBoolean   StorageType = "boolean"
	StorageDate      StorageType = "date"
	StorageTimestamp StorageType = "timestamp"
	StorageUUID      StorageType = "uuid"
	StorageEnum      StorageType = "enum"
	StorageRelation  StorageType = "relation"
	StorageTextList  StorageType = "text_list"
	StorageUUIDList  StorageType = "uuid_list"
	StorageJSON      StorageType = "json"
)

const (
	FeatureBaseColumn              StorageFeature = "base_column"
	FeatureBTreeIndex              StorageFeature = "btree_index"
	FeatureFoldedCompanion         StorageFeature = "folded_companion"
	FeatureAccentCompanion         StorageFeature = "accent_companion"
	FeatureTrigramIndex            StorageFeature = "trigram_index"
	FeatureGINIndex                StorageFeature = "gin_index"
	FeatureRelationEdgeTable       StorageFeature = "relation_edge_table"
	FeatureRelationSlotConstraints StorageFeature = "relation_slot_constraints"
)
