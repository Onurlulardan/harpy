package domain

import "time"

type ActiveUsers struct {
	DAU, WAU, MAU, ConcurrentUsers, TotalRequests int64
}

type TopUser struct {
	SubjectID    string
	RequestCount int64
	UniqueRoutes int64
	LastActive   time.Time
}

type RouteUsage struct {
	RoutePattern  string
	Method        string
	Feature       string
	RequestCount  int64
	AverageMillis float64
}

type DailyActivity struct {
	Date         string
	RequestCount int64
}

type UserSummary struct {
	SubjectID      string
	LastActive     time.Time
	TotalRequests  int64
	UniqueRoutes   int64
	TopRoutes      []RouteUsage
	DailyBreakdown []DailyActivity
}

type TimelineEntry struct {
	ID, RoutePattern, Method, Feature, Level string
	StatusCode                               int
	DurationMillis                           float64
	OccurredAt                               time.Time
}

type TimelinePage struct {
	Items                   []TimelineEntry
	Total                   int64
	Page, Limit, TotalPages int
}

type FeatureUsage struct {
	RoutePattern, Method, Feature string
	RequestCount, UniqueUsers     int64
	AverageMillis, ErrorRate      float64
	ErrorCount                    int64
}

type HeatmapCell struct {
	Day, Hour int
	Count     int64
}

type TrendPoint struct {
	Bucket      time.Time
	Requests    int64
	UniqueUsers int64
}

type Overview struct {
	ActiveUsers ActiveUsers
	TopFeatures []FeatureUsage
	TopUsers    []TopUser
	RecentTrend []TrendPoint
}
