package systemtransfer

import (
	"strings"
	"testing"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

func TestSourceCursorResumesAtDurableNextRow(t *testing.T) {
	reader, err := tabulartransfer.NewReader(tabulartransfer.FormatJSONL,
		strings.NewReader("{\"email\":\"one@test\"}\n{\"email\":\"two@test\"}\n"),
		tabulartransfer.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := NewSourceCursor(reader, tabulartransfer.FormatJSONL)
	if err != nil {
		t.Fatal(err)
	}
	if err = cursor.SkipProcessed(1); err != nil {
		t.Fatal(err)
	}
	rows, eof, err := cursor.NextBatch()
	if err != nil || !eof || len(rows) != 1 || rows[0].Index != 2 {
		t.Fatalf("rows=%+v eof=%t err=%v", rows, eof, err)
	}
}

func TestSourceCursorRejectsResumePastEOF(t *testing.T) {
	reader, err := tabulartransfer.NewReader(tabulartransfer.FormatJSONL,
		strings.NewReader("{\"email\":\"one@test\"}\n"),
		tabulartransfer.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	cursor, _ := NewSourceCursor(reader, tabulartransfer.FormatJSONL)
	if err = cursor.SkipProcessed(2); err != ErrWorkerState {
		t.Fatalf("skip error=%v", err)
	}
}
