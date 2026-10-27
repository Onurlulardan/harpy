package dashboardexport

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type rendererExecutor struct{ rows int64 }

func (executor rendererExecutor) Execute(
	_ context.Context, _ Pin, widget PinnedWidget,
) (WidgetResult, error) {
	return WidgetResult{Body: []byte(`{"z":2,"a":"` + widget.WidgetID.String() + `"}`),
		Rows: executor.rows}, nil
}

func TestDeterministicRendererFormatsAreStable(t *testing.T) {
	t.Parallel()
	for _, format := range []Format{FormatCSV, FormatXLSX, FormatPDF} {
		format := format
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()
			renderer, err := NewDeterministicRenderer(rendererExecutor{rows: 2})
			if err != nil {
				t.Fatal(err)
			}
			pin := rendererPin(format)
			var first, second bytes.Buffer
			firstStats, firstErr := renderer.Render(context.Background(), pin, &first)
			secondStats, secondErr := renderer.Render(context.Background(), pin, &second)
			if firstErr != nil || secondErr != nil || firstStats.Rows != 4 ||
				secondStats.Rows != 4 || !bytes.Equal(first.Bytes(), second.Bytes()) {
				t.Fatalf("render = %d/%d stable=%v errors=%v/%v",
					firstStats.Rows, secondStats.Rows,
					bytes.Equal(first.Bytes(), second.Bytes()), firstErr, secondErr)
			}
			signature := map[Format][]byte{FormatCSV: []byte("widget_id,"),
				FormatXLSX: []byte("PK"), FormatPDF: []byte("%PDF-1.7")}[format]
			if !bytes.HasPrefix(first.Bytes(), signature) {
				t.Fatalf("%s signature = %q", format, first.Bytes()[:min(12, first.Len())])
			}
		})
	}
}

func TestDeterministicRendererRejectsAggregateRowOverflow(t *testing.T) {
	t.Parallel()
	renderer, err := NewDeterministicRenderer(rendererExecutor{rows: 2})
	if err != nil {
		t.Fatal(err)
	}
	pin := rendererPin(FormatCSV)
	pin.MaximumRows = 3
	if _, err = renderer.Render(context.Background(), pin, &bytes.Buffer{}); !errors.Is(
		err, ErrUnavailable,
	) {
		t.Fatalf("Render() error = %v, want ErrUnavailable", err)
	}
}

func rendererPin(format Format) Pin {
	return Pin{TenantID: uuid.MustParse("019b1000-0000-7000-8000-000000000001"),
		ActorID:            uuid.MustParse("019b1000-0000-7000-8000-000000000002"),
		ExportID:           uuid.MustParse("019b1000-0000-7000-8000-000000000003"),
		DashboardID:        uuid.MustParse("019b1000-0000-7000-8000-000000000004"),
		DashboardVersionID: uuid.MustParse("019b1000-0000-7000-8000-000000000005"),
		Format:             format, Parameters: []byte(`{}`), MaximumRows: 10,
		MaximumPlaintextBytes: 1 << 20, ExpiresAt: time.Now().UTC().Add(time.Hour),
		Widgets: []PinnedWidget{
			{WidgetID: uuid.MustParse("019b1000-0000-7000-8000-000000000006"),
				QueryVersionID: uuid.MustParse("019b1000-0000-7000-8000-000000000007")},
			{WidgetID: uuid.MustParse("019b1000-0000-7000-8000-000000000008"),
				QueryVersionID: uuid.MustParse("019b1000-0000-7000-8000-000000000009"),
				Position:       1}}}
}
