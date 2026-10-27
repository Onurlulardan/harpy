package audithttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

func listFilter(request *http.Request) (domain.Filter, error) {
	query := request.URL.Query()
	if !onlyQueryKeys(query, "page", "limit", "owner_module", "entity_type", "entity_id",
		"form_slug", "action", "user_id", "org_id", "correlation_id", "start_time",
		"end_time", "search", "sort_by", "sort_direction") {
		return domain.Filter{}, application.ErrInvalid
	}
	filter, err := baseFilter(request, 20)
	if err != nil {
		return domain.Filter{}, err
	}
	filter.OwnerModule, filter.EntityType = query.Get("owner_module"), query.Get("entity_type")
	filter.EntityID, filter.FormSlug = query.Get("entity_id"), query.Get("form_slug")
	filter.Action, filter.Search = domain.Action(query.Get("action")), query.Get("search")
	if filter.ActorID, err = optionalUUID(query.Get("user_id")); err != nil {
		return domain.Filter{}, err
	}
	if filter.OrganizationID, err = optionalUUID(query.Get("org_id")); err != nil {
		return domain.Filter{}, err
	}
	if filter.CorrelationID, err = optionalUUID(query.Get("correlation_id")); err != nil {
		return domain.Filter{}, err
	}
	if filter.Start, err = optionalTime(query.Get("start_time")); err != nil {
		return domain.Filter{}, err
	}
	if filter.End, err = optionalTime(query.Get("end_time")); err != nil {
		return domain.Filter{}, err
	}
	if sort := query.Get("sort_by"); sort != "" && sort != "created_at" {
		return domain.Filter{}, application.ErrInvalid
	}
	if direction := query.Get("sort_direction"); direction != "" && direction != "DESC" {
		return domain.Filter{}, application.ErrInvalid
	}
	if filter.Validate() != nil {
		return domain.Filter{}, application.ErrInvalid
	}
	return filter, nil
}

func historyFilter(request *http.Request) (domain.Filter, error) {
	if !onlyQueryKeys(request.URL.Query(), "page", "limit") {
		return domain.Filter{}, application.ErrInvalid
	}
	return baseFilter(request, 50)
}

func baseFilter(request *http.Request, defaultLimit int) (domain.Filter, error) {
	page, limit, err := 1, defaultLimit, error(nil)
	if value := request.URL.Query().Get("page"); value != "" {
		page, err = positiveDecimal(value, 1000)
	}
	if err == nil {
		if value := request.URL.Query().Get("limit"); value != "" {
			limit, err = positiveDecimal(value, 100)
		}
	}
	filter := domain.Filter{Page: page, Limit: limit}
	if err != nil || filter.Validate() != nil {
		return domain.Filter{}, application.ErrInvalid
	}
	return filter, nil
}

func positiveDecimal(value string, maximum int) (int, error) {
	if value == "" || value[0] == '0' || len(value) > 4 {
		return 0, application.ErrInvalid
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > maximum {
		return 0, application.ErrInvalid
	}
	return parsed, nil
}

func optionalUUID(value string) (*domain.ID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := domain.ParseID(value)
	if err != nil {
		return nil, application.ErrInvalid
	}
	return &id, nil
}

func exactUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil || id.String() != value {
		return uuid.Nil, application.ErrInvalid
	}
	return id, nil
}

func exactEntryUUID(value string) (uuid.UUID, error) {
	id, err := exactUUID(value)
	if err != nil || id.Version() != 7 {
		return uuid.Nil, application.ErrInvalid
	}
	return id, nil
}

func optionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Nanosecond()%1000 != 0 {
		return nil, application.ErrInvalid
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func onlyQueryKeys(query map[string][]string, allowed ...string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		set[key] = struct{}{}
	}
	for key, values := range query {
		if _, found := set[key]; !found || len(values) != 1 {
			return false
		}
	}
	return true
}
