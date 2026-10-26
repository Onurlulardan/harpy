package platformops

import (
	"sort"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type traceTreeNode struct {
	view     spanTreeView
	children []*traceTreeNode
}

func traceDetailAPI(spans []platformopsapi.TraceSpan) traceDetailView {
	base := spans[0].StartedAt
	result := traceDetailView{CorrelationID: spans[0].CorrelationID,
		Spans: make([]traceSpanView, len(spans)), TotalSpans: len(spans)}
	latest := spans[0].EndedAt
	for index, span := range spans {
		result.Spans[index] = traceSpanAPI(span)
		if span.StartedAt.Before(base) {
			base = span.StartedAt
		}
		if span.EndedAt.After(latest) {
			latest = span.EndedAt
		}
	}
	result.TotalDurationMS = float64(latest.Sub(base).Microseconds()) / 1000
	result.Tree = buildTraceTree(spans, base)
	return result
}

func buildTraceTree(spans []platformopsapi.TraceSpan, base time.Time) spanTreeView {
	nodes := make(map[string]*traceTreeNode, len(spans))
	for _, span := range spans {
		nodes[span.SpanID] = &traceTreeNode{view: spanTreeView{
			SpanID: span.SpanID, ServiceName: span.ServiceName, Operation: span.Operation,
			SpanKind: span.SpanKind, DurationMS: span.DurationMillis,
			StatusCode: span.StatusCode, Level: span.Level,
			HasError:      span.StatusCode >= 500 || span.Level == "error" || span.Level == "fatal",
			StartOffsetMS: float64(span.StartedAt.Sub(base).Microseconds()) / 1000,
			Children:      make([]spanTreeView, 0)}}
	}
	var roots []*traceTreeNode
	for _, span := range spans {
		node := nodes[span.SpanID]
		parent := nodes[span.ParentSpanID]
		if span.ParentSpanID == "" || parent == nil {
			roots = append(roots, node)
		} else {
			parent.children = append(parent.children, node)
		}
	}
	sort.Slice(roots, func(left, right int) bool {
		return roots[left].view.StartOffsetMS < roots[right].view.StartOffsetMS
	})
	return materializeTraceTree(roots[0])
}

func materializeTraceTree(node *traceTreeNode) spanTreeView {
	sort.Slice(node.children, func(left, right int) bool {
		return node.children[left].view.StartOffsetMS < node.children[right].view.StartOffsetMS
	})
	result := node.view
	for _, child := range node.children {
		result.Children = append(result.Children, materializeTraceTree(child))
	}
	return result
}
