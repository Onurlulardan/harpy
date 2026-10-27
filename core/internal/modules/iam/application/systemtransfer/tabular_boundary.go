package systemtransfer

import tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"

type ExportCell = tabulartransfer.Cell
type ExportRowSource = tabulartransfer.RowSource

const (
	ExportCellNull    = tabulartransfer.CellNull
	ExportCellText    = tabulartransfer.CellText
	ExportCellBoolean = tabulartransfer.CellBoolean
	ExportCellJSON    = tabulartransfer.CellJSON
)
