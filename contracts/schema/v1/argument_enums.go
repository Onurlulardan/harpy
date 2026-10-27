package schemav1

type StorageType string

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

type Cardinality string

const (
	CardinalityOne  Cardinality = "one"
	CardinalityMany Cardinality = "many"
)

type DeletePolicy string

const (
	DeleteRestrict DeletePolicy = "restrict"
	DeleteCascade  DeletePolicy = "cascade"
)

type IndexMethod string

const (
	IndexBTree IndexMethod = "btree"
	IndexGIN   IndexMethod = "gin"
)

type ConstraintKind string

const (
	ConstraintNotNull    ConstraintKind = "not_null"
	ConstraintUnique     ConstraintKind = "unique"
	ConstraintTextLength ConstraintKind = "text_length"
)
