package dashboardexport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

type WidgetResult struct {
	Body []byte
	Rows int64
}

type WidgetExecutor interface {
	Execute(context.Context, Pin, PinnedWidget) (WidgetResult, error)
}

type RenderStats struct{ Rows int64 }

type Renderer interface {
	Render(context.Context, Pin, io.Writer) (RenderStats, error)
}

type DeterministicRenderer struct{ executor WidgetExecutor }

func NewDeterministicRenderer(executor WidgetExecutor) (*DeterministicRenderer, error) {
	if nilPort(executor) {
		return nil, fmt.Errorf("analytics dashboard export executor is required")
	}
	return &DeterministicRenderer{executor: executor}, nil
}

type renderedWidget struct {
	PinnedWidget
	Body []byte
	Rows int64
}

func (renderer *DeterministicRenderer) Render(
	ctx context.Context, pin Pin, destination io.Writer,
) (RenderStats, error) {
	if renderer == nil || ctx == nil || ctx.Err() != nil || destination == nil ||
		!pin.Format.Valid() || len(pin.Widgets) < 1 || len(pin.Widgets) > MaximumWidgets {
		return RenderStats{}, ErrInvalid
	}
	items := make([]renderedWidget, len(pin.Widgets))
	var rows int64
	for index, widget := range pin.Widgets {
		result, err := renderer.executor.Execute(ctx, pin, widget)
		canonical, canonicalErr := canonicaljsonv1.Canonicalize(result.Body, 1<<20)
		if err != nil || canonicalErr != nil || result.Rows < 0 ||
			rows > pin.MaximumRows-result.Rows || !json.Valid(canonical) {
			return RenderStats{}, ErrUnavailable
		}
		rows += result.Rows
		items[index] = renderedWidget{PinnedWidget: widget,
			Body: canonical, Rows: result.Rows}
	}
	var err error
	if pin.Format == FormatPDF {
		err = writePDF(destination, pin, items)
	} else {
		err = writeTabular(destination, pin.Format, items)
	}
	if err != nil {
		return RenderStats{}, err
	}
	return RenderStats{Rows: rows}, nil
}

func writeTabular(destination io.Writer, format Format, items []renderedWidget) error {
	transferFormat := tabulartransfer.FormatCSV
	if format == FormatXLSX {
		transferFormat = tabulartransfer.FormatXLSX
	}
	_, err := tabulartransfer.Write(destination, transferFormat,
		[]string{"widget_id", "query_version_id", "rows", "result_json"},
		&widgetRows{items: items})
	return err
}

type widgetRows struct {
	items []renderedWidget
	index int
}

func (source *widgetRows) Next() ([]tabulartransfer.Cell, error) {
	if source.index >= len(source.items) {
		return nil, io.EOF
	}
	item := source.items[source.index]
	source.index++
	return []tabulartransfer.Cell{
		{Kind: tabulartransfer.CellText, Text: item.WidgetID.String()},
		{Kind: tabulartransfer.CellText, Text: item.QueryVersionID.String()},
		{Kind: tabulartransfer.CellNumber, Text: fmt.Sprintf("%d", item.Rows)},
		{Kind: tabulartransfer.CellJSON, Text: string(item.Body)},
	}, nil
}
