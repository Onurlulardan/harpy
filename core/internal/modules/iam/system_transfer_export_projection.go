package iam

import systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"

const systemTransferExportProjectionSize = 49

type systemTransferExportProjection struct {
	fields []string
	cursor string
	more   bool
}

func systemTransferExportProjections(
	columns []string,
) ([]systemTransferExportProjection, error) {
	if len(columns) == 0 {
		return nil, systemtransfer.ErrExportQueryInvalid
	}
	nonID := make([]string, 0, len(columns))
	for _, column := range columns {
		if column != "id" {
			nonID = append(nonID, column)
		}
	}
	if len(nonID) == 0 {
		return []systemTransferExportProjection{{fields: []string{"id"}}}, nil
	}
	result := make([]systemTransferExportProjection, 0,
		(len(nonID)+systemTransferExportProjectionSize-1)/systemTransferExportProjectionSize)
	for start := 0; start < len(nonID); start += systemTransferExportProjectionSize {
		end := min(start+systemTransferExportProjectionSize, len(nonID))
		fields := append([]string(nil), nonID[start:end]...)
		fields = append(fields, "id")
		result = append(result, systemTransferExportProjection{fields: fields})
	}
	return result, nil
}

func (projection *systemTransferExportProjection) after() *string {
	if projection == nil || !projection.more {
		return nil
	}
	value := projection.cursor
	return &value
}
