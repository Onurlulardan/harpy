package postgres

import "time"

type activeDocument struct {
	DAU             int64 `json:"dau"`
	WAU             int64 `json:"wau"`
	MAU             int64 `json:"mau"`
	ConcurrentUsers int64 `json:"concurrent_users"`
	TotalRequests   int64 `json:"total_requests_today"`
}

type topUserDocument struct {
	SubjectID    string    `json:"subject_id"`
	RequestCount int64     `json:"request_count"`
	UniqueRoutes int64     `json:"unique_routes"`
	LastActive   time.Time `json:"last_active"`
}

type routeDocument struct {
	RoutePattern  string  `json:"route_pattern"`
	Method        string  `json:"method"`
	RequestCount  int64   `json:"request_count"`
	AverageMillis float64 `json:"average_millis"`
}

type dailyDocument struct {
	Date         string `json:"date"`
	RequestCount int64  `json:"request_count"`
}

type summaryDocument struct {
	SubjectID      string          `json:"subject_id"`
	LastActive     time.Time       `json:"last_active"`
	TotalRequests  int64           `json:"total_requests"`
	UniqueRoutes   int64           `json:"unique_routes"`
	TopRoutes      []routeDocument `json:"top_routes"`
	DailyBreakdown []dailyDocument `json:"daily_breakdown"`
}

type timelineEntryDocument struct {
	ID             string    `json:"id"`
	RoutePattern   string    `json:"route_pattern"`
	Method         string    `json:"method"`
	StatusCode     int       `json:"status_code"`
	DurationMillis float64   `json:"duration_millis"`
	OccurredAt     time.Time `json:"occurred_at"`
}

type timelineDocument struct {
	Items      []timelineEntryDocument `json:"items"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

type featureDocument struct {
	RoutePattern  string  `json:"route_pattern"`
	Method        string  `json:"method"`
	RequestCount  int64   `json:"request_count"`
	UniqueUsers   int64   `json:"unique_users"`
	AverageMillis float64 `json:"average_millis"`
	ErrorCount    int64   `json:"error_count"`
	ErrorRate     float64 `json:"error_rate"`
}

type heatmapDocument struct {
	Day   int   `json:"day"`
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type trendDocument struct {
	Bucket      time.Time `json:"bucket"`
	Requests    int64     `json:"requests"`
	UniqueUsers int64     `json:"unique_users"`
}

type overviewDocument struct {
	ActiveUsers activeDocument    `json:"active_users"`
	TopFeatures []featureDocument `json:"top_features"`
	TopUsers    []topUserDocument `json:"top_users"`
	RecentTrend []trendDocument   `json:"recent_trend"`
}
