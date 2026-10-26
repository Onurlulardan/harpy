package postgres

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func decodeFeatures(document []byte) ([]domain.FeatureUsage, error) {
	var wire []featureDocument
	if decodeDocument(document, &wire) != nil || len(wire) > 100 {
		return nil, domain.ErrUnavailable
	}
	result := make([]domain.FeatureUsage, 0, len(wire))
	for _, item := range wire {
		if !domain.ValidRoutePattern(item.RoutePattern) || item.Method == "" ||
			!validCount(item.RequestCount, item.UniqueUsers, item.ErrorCount) ||
			!validMetric(item.AverageMillis) || !validMetric(item.ErrorRate) ||
			item.ErrorRate > 100 {
			return nil, domain.ErrUnavailable
		}
		result = append(result, domain.FeatureUsage{RoutePattern: item.RoutePattern,
			Method: item.Method, RequestCount: item.RequestCount,
			UniqueUsers: item.UniqueUsers, AverageMillis: item.AverageMillis,
			ErrorCount: item.ErrorCount, ErrorRate: item.ErrorRate})
	}
	return result, nil
}

func decodeHeatmap(document []byte) ([]domain.HeatmapCell, error) {
	var wire []heatmapDocument
	if decodeDocument(document, &wire) != nil || len(wire) > 168 {
		return nil, domain.ErrUnavailable
	}
	result := make([]domain.HeatmapCell, 0, len(wire))
	for _, item := range wire {
		if item.Day < 0 || item.Day > 6 || item.Hour < 0 || item.Hour > 23 ||
			!validCount(item.Count) {
			return nil, domain.ErrUnavailable
		}
		result = append(result, domain.HeatmapCell{Day: item.Day, Hour: item.Hour,
			Count: item.Count})
	}
	return result, nil
}

func decodeTrend(document []byte) ([]domain.TrendPoint, error) {
	var wire []trendDocument
	if decodeDocument(document, &wire) != nil || len(wire) > 745 {
		return nil, domain.ErrUnavailable
	}
	result := make([]domain.TrendPoint, 0, len(wire))
	for _, item := range wire {
		if !validTime(item.Bucket) || !validCount(item.Requests, item.UniqueUsers) {
			return nil, domain.ErrUnavailable
		}
		result = append(result, domain.TrendPoint{Bucket: item.Bucket.UTC(),
			Requests: item.Requests, UniqueUsers: item.UniqueUsers})
	}
	return result, nil
}

func decodeOverview(document []byte, scope domain.Scope) (domain.Overview, error) {
	var wire overviewDocument
	if decodeDocument(document, &wire) != nil {
		return domain.Overview{}, domain.ErrUnavailable
	}
	activeJSON, _ := json.Marshal(wire.ActiveUsers)
	featuresJSON, _ := json.Marshal(wire.TopFeatures)
	usersJSON, _ := json.Marshal(wire.TopUsers)
	trendJSON, _ := json.Marshal(wire.RecentTrend)
	active, activeErr := decodeActiveUsers(activeJSON)
	features, featureErr := decodeFeatures(featuresJSON)
	users, userErr := decodeTopUsers(usersJSON, scope)
	trend, trendErr := decodeTrend(trendJSON)
	if activeErr != nil || featureErr != nil || userErr != nil || trendErr != nil {
		return domain.Overview{}, domain.ErrUnavailable
	}
	return domain.Overview{ActiveUsers: active, TopFeatures: features,
		TopUsers: users, RecentTrend: trend}, nil
}
