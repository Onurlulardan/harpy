package usagehttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func timeRange(request *http.Request, extras ...string) (application.TimeRangeRequest, error) {
	allowed := append([]string{"start_time", "end_time"}, extras...)
	if request == nil || !onlyQueryKeys(request, allowed...) {
		return application.TimeRangeRequest{}, domain.ErrInvalid
	}
	start, startErr := optionalTime(request.URL.Query().Get("start_time"))
	end, endErr := optionalTime(request.URL.Query().Get("end_time"))
	if startErr != nil || endErr != nil || (start == nil) != (end == nil) {
		return application.TimeRangeRequest{}, domain.ErrInvalid
	}
	return application.TimeRangeRequest{Start: start, End: end}, nil
}

func optionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Nanosecond()%1000 != 0 {
		return nil, domain.ErrInvalid
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func positiveQuery(request *http.Request, key string, fallback, maximum int) (int, error) {
	value := request.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	if value[0] == '0' || len(value) > 4 {
		return 0, domain.ErrInvalid
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > maximum {
		return 0, domain.ErrInvalid
	}
	return parsed, nil
}

func onlyQueryKeys(request *http.Request, allowed ...string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		set[key] = struct{}{}
	}
	for key, values := range request.URL.Query() {
		if _, ok := set[key]; !ok || len(values) != 1 {
			return false
		}
	}
	return true
}
