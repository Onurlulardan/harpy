package platformops

import (
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type traceSummaryView struct {
	CorrelationID   string    `json:"correlation_id"`
	FirstSpan       time.Time `json:"first_span"`
	LastSpan        time.Time `json:"last_span"`
	TotalDurationMS float64   `json:"total_duration_ms"`
	SpanCount       int       `json:"span_count"`
	Services        []string  `json:"services"`
	RootOperation   string    `json:"root_operation"`
	RootStatusCode  int       `json:"root_status_code,omitempty"`
	HasErrors       bool      `json:"has_errors"`
}

type tracePageView struct {
	Traces     []traceSummaryView `json:"traces"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

type traceSpanView struct {
	ID            string    `json:"id"`
	CorrelationID string    `json:"correlation_id"`
	SpanID        string    `json:"span_id"`
	ParentSpanID  string    `json:"parent_span_id,omitempty"`
	ServiceName   string    `json:"service_name"`
	Operation     string    `json:"operation"`
	SpanKind      string    `json:"span_kind"`
	Method        string    `json:"method,omitempty"`
	Path          string    `json:"path,omitempty"`
	StatusCode    int       `json:"status_code,omitempty"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	DurationMS    float64   `json:"duration_ms"`
	Level         string    `json:"level"`
	Message       string    `json:"message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func tracePageAPI(page platformopsapi.TracePage) tracePageView {
	result := tracePageView{Traces: make([]traceSummaryView, len(page.Traces)),
		Total: page.Total, Page: page.Page, Limit: page.Limit, TotalPages: page.TotalPages}
	for index, item := range page.Traces {
		result.Traces[index] = traceSummaryView{CorrelationID: item.CorrelationID,
			FirstSpan: item.FirstSpan, LastSpan: item.LastSpan,
			TotalDurationMS: item.TotalDurationMillis, SpanCount: item.SpanCount,
			Services: item.Services, RootOperation: item.RootOperation,
			RootStatusCode: item.RootStatusCode, HasErrors: item.HasErrors}
	}
	return result
}
