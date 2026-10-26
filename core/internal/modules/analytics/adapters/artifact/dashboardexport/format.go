package dashboardexportartifact

import (
	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func artifactFormat(format exportapp.Format) (fcartv1.Format, error) {
	switch format {
	case exportapp.FormatCSV:
		return fcartv1.FormatCSV, nil
	case exportapp.FormatXLSX:
		return fcartv1.FormatXLSX, nil
	case exportapp.FormatPDF:
		return fcartv1.FormatPDF, nil
	default:
		return "", exportapp.ErrInvalid
	}
}
