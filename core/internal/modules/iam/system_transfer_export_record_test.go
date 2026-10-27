package iam

import (
	"testing"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestSystemTransferExportRecordAcceptsAdoptedCanonicalUUID(t *testing.T) {
	record, err := decodeSystemTransferExportRecord(
		[]byte(`{"id":"11111111-1111-4111-8111-111111111111","enabled":true}`),
		[]string{"id", "enabled"},
	)
	if err != nil || record.id.String() != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("record=%#v err=%v", record, err)
	}
	cell, err := systemTransferExportCell(
		queryapp.ExportLogicalTypeBoolean, record.members["enabled"],
	)
	if err != nil || cell.Kind != systemtransfer.ExportCellBoolean || cell.Text != "true" {
		t.Fatalf("cell=%#v err=%v", cell, err)
	}
}

func TestSystemTransferExportProjectionKeepsStableIDOutOfOutputLimit(t *testing.T) {
	columns := make([]string, 50)
	for index := range columns {
		columns[index] = "field_" + string(rune('a'+index))
	}
	projections, err := systemTransferExportProjections(columns)
	if err != nil || len(projections) != 2 ||
		len(projections[0].fields) != 50 ||
		projections[0].fields[49] != "id" || projections[1].fields[1] != "id" {
		t.Fatalf("projections=%#v err=%v", projections, err)
	}
}
