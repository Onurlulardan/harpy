package fcartv1

type Format string

const (
	FormatCSV   Format = "csv"
	FormatJSONL Format = "jsonl"
	FormatXLSX  Format = "xlsx"
	FormatPDF   Format = "pdf"
)

func (format Format) Validate() error {
	switch format {
	case FormatCSV, FormatJSONL, FormatXLSX, FormatPDF:
		return nil
	default:
		return ErrInvalidFormat
	}
}

func (format Format) wireCode() (byte, error) {
	switch format {
	case FormatCSV:
		return 1, nil
	case FormatJSONL:
		return 2, nil
	case FormatXLSX:
		return 3, nil
	case FormatPDF:
		return 4, nil
	default:
		return 0, ErrInvalidFormat
	}
}

func formatFromWire(code byte) (Format, error) {
	switch code {
	case 1:
		return FormatCSV, nil
	case 2:
		return FormatJSONL, nil
	case 3:
		return FormatXLSX, nil
	case 4:
		return FormatPDF, nil
	default:
		return "", ErrInvalidFormat
	}
}
