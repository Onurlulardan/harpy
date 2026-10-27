package dashboardexport

type Format string

const (
	FormatCSV  Format = "csv"
	FormatXLSX Format = "xlsx"
	FormatPDF  Format = "pdf"
)

func (format Format) Valid() bool {
	return format == FormatCSV || format == FormatXLSX || format == FormatPDF
}

func (format Format) ContentType() string {
	switch format {
	case FormatCSV:
		return "text/csv; charset=utf-8"
	case FormatXLSX:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case FormatPDF:
		return "application/pdf"
	default:
		return ""
	}
}

func (format Format) Extension() string {
	if !format.Valid() {
		return ""
	}
	return string(format)
}
