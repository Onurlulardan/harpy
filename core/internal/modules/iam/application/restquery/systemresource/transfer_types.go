package systemresourcequery

import restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

type ExportQuery = restquery.QueryRequest
type ExportLogicalType = restquery.LogicalType
type ExportFieldCatalog = restquery.FieldCatalog
type ExportFilter = restquery.RequestFilter

const (
	ExportQueryModeRecord     = restquery.QueryModeRecord
	ExportCountModeNone       = restquery.CountModeNone
	ExportLogicalTypeBoolean  = restquery.LogicalTypeBoolean
	ExportLogicalTypeJSON     = restquery.LogicalTypeJSON
	ExportLogicalTypeTextList = restquery.LogicalTypeTextList
	ExportLogicalTypeUUIDList = restquery.LogicalTypeUUIDList
)
