package systemtransferartifact

import (
	"strings"
	"testing"
)

func TestAnalyzeJSONLUsesDeterministicColumnsAndMixedKinds(t *testing.T) {
	analysis, err := analyzeSource("jsonl", []byte(
		"{\"z\":1,\"a\":true,\"nested\":{\"id\":1}}\n"+
			"{\"z\":\"one\",\"a\":false,\"items\":[1]}\n"), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Columns) != 4 || analysis.Columns[0].Name != "a" ||
		analysis.Columns[1].Name != "items" || analysis.Columns[2].Name != "nested" ||
		analysis.Columns[3].Name != "z" || analysis.Columns[3].Kind != "mixed" ||
		len(analysis.Samples) != 2 {
		t.Fatalf("unexpected analysis: %#v", analysis)
	}
}

func TestAnalyzeTruncatedCaptureIgnoresOnlyIncompleteTail(t *testing.T) {
	analysis, err := analyzeSource("csv", []byte("email,active\na@b.test,true\npartial"), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Columns) != 2 || len(analysis.Samples) != 1 ||
		analysis.Samples[0]["email"] != "a@b.test" {
		t.Fatalf("unexpected bounded analysis: %#v", analysis)
	}
}

func TestAnalyzeOmitsOversizedSampleWithoutRejectingColumns(t *testing.T) {
	value := "payload\n" + strings.Repeat("x", maximumAnalysisSampleBytes+1) + "\n"
	analysis, err := analyzeSource("csv", []byte(value), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Columns) != 1 || len(analysis.Samples) != 0 {
		t.Fatalf("unexpected oversized sample analysis: %#v", analysis)
	}
}
