package postgres

import (
	"errors"
	"testing"
)

var errUnexpectedScan = errors.New("unexpected scan destination")

type contractRow struct {
	values []any
}

func (row contractRow) Scan(destinations ...any) error {
	for index, destination := range destinations {
		switch target := destination.(type) {
		case *int64:
			*target = row.values[index].(int64)
		case *string:
			*target = row.values[index].(string)
		case *bool:
			*target = row.values[index].(bool)
		default:
			return errUnexpectedScan
		}
	}
	return nil
}

func TestEmbeddedContractMakesBaselineProbeReady(t *testing.T) {
	t.Parallel()
	contract, err := loadEmbeddedBinaryContract()
	if err != nil {
		t.Fatal(err)
	}
	result, err := probeSchemaContractRow(contractRow{values: []any{
		int64(1), int64(1),
		contract.MigrationManifestSHA256, contract.ExtensionManifestSHA256,
		"ready", "33333333-3333-4333-8333-333333333333", int64(1), true,
	}}, contract)
	if err != nil || !result.Ready() || result.Reason() != SchemaReady {
		t.Fatalf("baseline probe = %+v, %v", result, err)
	}
}
