package notificationhttp

import (
	"encoding/base64"
	"strconv"
	"strings"

	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func decodeInboxQuery(raw string) (ports.MessageFilter, error) {
	request, err := httpdecode.GET().Decode(raw)
	if err != nil || request.Mode() != restapi.QueryModeRecord ||
		len(request.Select()) != 0 || len(request.Includes()) != 0 ||
		request.Count() != restapi.CountModeNone {
		return ports.MessageFilter{}, errQueryInvalid
	}
	if _, present := request.Search(); present {
		return ports.MessageFilter{}, errQueryInvalid
	}
	if _, present := request.Aggregate(); present {
		return ports.MessageFilter{}, errQueryInvalid
	}
	if !validInboxSort(request.Sort()) || request.Page().Size() > 100 {
		return ports.MessageFilter{}, errQueryInvalid
	}
	filter := ports.MessageFilter{Limit: request.Page().Size()}
	if cursor, present := request.Page().After(); present {
		filter.Before, err = decodeInboxCursor(cursor)
		if err != nil {
			return ports.MessageFilter{}, errQueryInvalid
		}
	}
	where, present := request.Where()
	if present && !mapInboxWhere(where, &filter) {
		return ports.MessageFilter{}, errQueryInvalid
	}
	return filter, nil
}

func validInboxSort(sorts []restapi.RequestSort) bool {
	return len(sorts) == 0 || len(sorts) == 1 &&
		sorts[0].Field() == "inbox_revision" &&
		sorts[0].Direction() == restapi.DirectionDescending
}

func mapInboxWhere(where restapi.RequestFilter, target *ports.MessageFilter) bool {
	children := where.Children()
	if where.Kind() == restapi.RequestFilterPredicate {
		children = []restapi.RequestFilter{where}
	} else if where.Kind() != restapi.RequestFilterAnd {
		return false
	}
	seen := map[string]bool{}
	for _, child := range children {
		predicate, ok := child.Predicate()
		if !ok || predicate.Operator() != restapi.OperatorEq ||
			len(predicate.Values()) != 1 || seen[predicate.Field()] {
			return false
		}
		seen[predicate.Field()] = true
		value, ok, err := predicate.Values()[0].TextValue()
		if err != nil || !ok || !mapInboxPredicate(predicate.Field(), value, target) {
			return false
		}
	}
	return true
}

func mapInboxPredicate(field, value string, target *ports.MessageFilter) bool {
	switch field {
	case "unread":
		if value != "true" {
			return false
		}
		target.UnreadOnly = true
		return true
	case "category":
		category := domain.Category(value)
		if !category.Valid() {
			return false
		}
		target.Category = &category
		return true
	default:
		return false
	}
}

func encodeInboxCursor(raw string) (string, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 1 || strconv.FormatInt(value, 10) != raw {
		return "", errQueryInvalid
	}
	return "nr1." + base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

func decodeInboxCursor(value string) (string, error) {
	if !strings.HasPrefix(value, "nr1.") {
		return "", errQueryInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, "nr1."))
	if err != nil || len(decoded) == 0 || len(decoded) > 20 {
		return "", errQueryInvalid
	}
	raw := string(decoded)
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 1 || strconv.FormatInt(parsed, 10) != raw {
		return "", errQueryInvalid
	}
	return raw, nil
}
