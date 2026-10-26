package usagehttp

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type activeUsersView struct {
	DAU             int64 `json:"dau"`
	WAU             int64 `json:"wau"`
	MAU             int64 `json:"mau"`
	ConcurrentUsers int64 `json:"concurrent_users"`
	TotalRequests   int64 `json:"total_requests_today"`
}

type topUserView struct {
	UserID          string    `json:"user_id"`
	RequestCount    int64     `json:"request_count"`
	UniqueEndpoints int64     `json:"unique_endpoints"`
	LastActive      time.Time `json:"last_active"`
}

type pathUsageView struct {
	Path          string  `json:"path"`
	Method        string  `json:"method"`
	Feature       string  `json:"feature"`
	RequestCount  int64   `json:"request_count"`
	AverageMillis float64 `json:"avg_duration_ms"`
}

type dailyActivityView struct {
	Date         string `json:"date"`
	RequestCount int64  `json:"request_count"`
}

type userSummaryView struct {
	UserID          string              `json:"user_id"`
	TotalRequests   int64               `json:"total_requests"`
	UniqueEndpoints int64               `json:"unique_endpoints"`
	LastActive      time.Time           `json:"last_active"`
	TopPaths        []pathUsageView     `json:"top_paths"`
	DailyBreakdown  []dailyActivityView `json:"daily_breakdown"`
}

type timelineEntryView struct {
	ID         string    `json:"id"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	Feature    string    `json:"feature"`
	StatusCode int       `json:"status_code"`
	Duration   float64   `json:"duration_ms"`
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
}

type timelinePageView struct {
	Items      []timelineEntryView `json:"items"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

type featureView struct {
	Path          string  `json:"path"`
	Method        string  `json:"method"`
	Feature       string  `json:"feature"`
	RequestCount  int64   `json:"request_count"`
	UniqueUsers   int64   `json:"unique_users"`
	AverageMillis float64 `json:"avg_duration_ms"`
	ErrorCount    int64   `json:"error_count"`
	ErrorRate     float64 `json:"error_rate"`
}

type heatmapView struct {
	Day   int   `json:"day"`
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type trendView struct {
	Bucket      time.Time `json:"bucket"`
	Requests    int64     `json:"requests"`
	UniqueUsers int64     `json:"unique_users"`
}

type overviewView struct {
	ActiveUsers activeUsersView `json:"active_users"`
	TopFeatures []featureView   `json:"top_features"`
	TopUsers    []topUserView   `json:"top_users"`
	RecentTrend []trendView     `json:"recent_trend"`
}

func activeUsersViewOf(value domain.ActiveUsers) activeUsersView {
	return activeUsersView{value.DAU, value.WAU, value.MAU,
		value.ConcurrentUsers, value.TotalRequests}
}

func topUsersViewOf(values []domain.TopUser) []topUserView {
	result := make([]topUserView, 0, len(values))
	for _, value := range values {
		result = append(result, topUserView{value.SubjectID, value.RequestCount,
			value.UniqueRoutes, value.LastActive})
	}
	return result
}

func featureViewsOf(values []domain.FeatureUsage) []featureView {
	result := make([]featureView, 0, len(values))
	for _, value := range values {
		result = append(result, featureView{value.RoutePattern, value.Method, value.Feature,
			value.RequestCount, value.UniqueUsers, value.AverageMillis,
			value.ErrorCount, value.ErrorRate})
	}
	return result
}

func trendViewsOf(values []domain.TrendPoint) []trendView {
	result := make([]trendView, 0, len(values))
	for _, value := range values {
		result = append(result, trendView{value.Bucket, value.Requests, value.UniqueUsers})
	}
	return result
}

func overviewViewOf(value domain.Overview) overviewView {
	return overviewView{activeUsersViewOf(value.ActiveUsers), featureViewsOf(value.TopFeatures),
		topUsersViewOf(value.TopUsers), trendViewsOf(value.RecentTrend)}
}
