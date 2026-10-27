package api

import "fmt"

const (
	MaximumUploadBytes       int64 = 512 << 20
	MaximumDecompressedBytes int64 = 2 << 30
	MaximumRows              int64 = 1_000_000
	MaximumColumns                 = 500
	MaximumRowBytes          int64 = 1 << 20
	MaximumPageRows                = 1_000
	MaximumMappedPageBytes   int64 = 8 << 20
	MaximumErrors                  = 10_000
	MaximumAnalysisRows            = 20
)

type Limits struct {
	UploadBytes, DecompressedBytes, RowBytes, MappedPageBytes int64
	Rows, Columns, PageRows, Errors, AnalysisRows             int
}

func DefaultLimits() Limits {
	return Limits{UploadBytes: MaximumUploadBytes,
		DecompressedBytes: MaximumDecompressedBytes, RowBytes: MaximumRowBytes,
		MappedPageBytes: MaximumMappedPageBytes, Rows: int(MaximumRows),
		Columns: MaximumColumns, PageRows: MaximumPageRows,
		Errors: MaximumErrors, AnalysisRows: MaximumAnalysisRows}
}

func (limits Limits) Validate() error {
	maximum := DefaultLimits()
	if limits.UploadBytes < 1 || limits.UploadBytes > maximum.UploadBytes ||
		limits.DecompressedBytes < 1 || limits.DecompressedBytes > maximum.DecompressedBytes ||
		limits.RowBytes < 1 || limits.RowBytes > maximum.RowBytes ||
		limits.MappedPageBytes < 1 || limits.MappedPageBytes > maximum.MappedPageBytes ||
		limits.Rows < 1 || limits.Rows > maximum.Rows || limits.Columns < 1 ||
		limits.Columns > maximum.Columns || limits.PageRows < 1 ||
		limits.PageRows > maximum.PageRows || limits.Errors < 1 ||
		limits.Errors > maximum.Errors || limits.AnalysisRows < 1 ||
		limits.AnalysisRows > maximum.AnalysisRows {
		return fmt.Errorf("tabular transfer limits are invalid")
	}
	return nil
}
