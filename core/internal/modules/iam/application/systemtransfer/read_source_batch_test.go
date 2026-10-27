package systemtransfer

import (
	"bytes"
	"strings"
	"testing"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

func TestReadSourceBatchConsumesSharedCSVAndJSONLReaders(t *testing.T) {
	tests := []struct {
		name, document, fragment string
		format                   tabulartransfer.Format
	}{
		{name: "csv", format: tabulartransfer.FormatCSV,
			document: "age,name\n42,Ada\n", fragment: `"age":"42"`},
		{name: "jsonl", format: tabulartransfer.FormatJSONL,
			document: `{"age":42,"name":"Ada"}` + "\n", fragment: `"age":42`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader, err := tabulartransfer.NewReader(test.format,
				bytes.NewBufferString(test.document), tabulartransfer.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			rows, done, err := ReadSourceBatch(reader, test.format, 2)
			if err != nil || !done || len(rows) != 1 ||
				!strings.Contains(string(rows[0].JSON), test.fragment) ||
				rows[0].Validate() != nil {
				t.Fatalf("rows=%+v done=%v err=%v", rows, done, err)
			}
		})
	}
}
