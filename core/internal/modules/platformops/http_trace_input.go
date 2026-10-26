package platformops

import (
	"net/http"
	"strconv"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

var traceQueryKeys = map[string]bool{
	"page": true, "limit": true, "service_name": true, "operation": true,
	"level": true, "status_code": true, "min_status": true, "max_status": true,
	"min_duration": true, "max_duration": true, "start_time": true,
	"end_time": true, "search": true, "sort_by": true, "sort_direction": true,
}

type traceFilterDocument struct {
	Page            int        `json:"page,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	ServiceName     string     `json:"service_name,omitempty"`
	Operation       string     `json:"operation,omitempty"`
	Level           string     `json:"level,omitempty"`
	Search          string     `json:"search,omitempty"`
	StatusCode      int        `json:"status_code,omitempty"`
	MinStatus       int        `json:"min_status,omitempty"`
	MaxStatus       int        `json:"max_status,omitempty"`
	MinimumDuration float64    `json:"min_duration,omitempty"`
	MaximumDuration float64    `json:"max_duration,omitempty"`
	StartTime       *time.Time `json:"start_time,omitempty"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	SortBy          string     `json:"sort_by,omitempty"`
	SortDirection   string     `json:"sort_direction,omitempty"`
}

func decodeTraceSearch(
	w http.ResponseWriter,
	r *http.Request,
) (platformopsapi.TraceFilter, error) {
	var document traceFilterDocument
	if err := httpinput.DecodeJSON(w, r, 8192, &document); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	return document.traceFilter(), nil
}

func traceFilterFromQuery(r *http.Request) (platformopsapi.TraceFilter, error) {
	if r == nil {
		return platformopsapi.TraceFilter{}, errHTTPRequestInvalid
	}
	query := r.URL.Query()
	for key, values := range query {
		if !traceQueryKeys[key] || len(values) != 1 {
			return platformopsapi.TraceFilter{}, errHTTPRequestInvalid
		}
	}
	document := traceFilterDocument{ServiceName: query.Get("service_name"),
		Operation: query.Get("operation"), Level: query.Get("level"),
		Search: query.Get("search"), SortBy: query.Get("sort_by"),
		SortDirection: query.Get("sort_direction")}
	var err error
	if document.Page, err = queryInt(query.Get("page")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.Limit, err = queryInt(query.Get("limit")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.StatusCode, err = queryInt(query.Get("status_code")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.MinStatus, err = queryInt(query.Get("min_status")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.MaxStatus, err = queryInt(query.Get("max_status")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.MinimumDuration, err = queryFloat(query.Get("min_duration")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.MaximumDuration, err = queryFloat(query.Get("max_duration")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.StartTime, err = queryTime(query.Get("start_time")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	if document.EndTime, err = queryTime(query.Get("end_time")); err != nil {
		return platformopsapi.TraceFilter{}, err
	}
	return document.traceFilter(), nil
}

func (document traceFilterDocument) traceFilter() platformopsapi.TraceFilter {
	if document.Page == 0 {
		document.Page = 1
	}
	if document.Limit == 0 {
		document.Limit = 50
	}
	if document.SortBy == "" {
		document.SortBy = "started_at"
	}
	if document.SortDirection == "" {
		document.SortDirection = "desc"
	}
	return platformopsapi.TraceFilter{Page: document.Page, Limit: document.Limit,
		ServiceName: document.ServiceName, Operation: document.Operation,
		Level: document.Level, Search: document.Search, StatusCode: document.StatusCode,
		MinStatus: document.MinStatus, MaxStatus: document.MaxStatus,
		MinimumDuration: document.MinimumDuration, MaximumDuration: document.MaximumDuration,
		StartedAfter: document.StartTime, EndedBefore: document.EndTime,
		SortBy: document.SortBy, SortDirection: document.SortDirection}
}

func queryInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errHTTPRequestInvalid
	}
	return parsed, nil
}

func queryFloat(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, errHTTPRequestInvalid
	}
	return parsed, nil
}

func queryTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, errHTTPRequestInvalid
	}
	return &parsed, nil
}
