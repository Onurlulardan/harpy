package adoption

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var (
	decimalSegment = regexp.MustCompile(`^[0-9]{1,20}$`)
	hexSegment     = regexp.MustCompile(`^[0-9a-fA-F]{16,}$`)
	ulidSegment    = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	opaqueSegment  = regexp.MustCompile(`^[A-Za-z0-9_-]{24,}$`)
	placeholder    = regexp.MustCompile(`^\{[a-z][a-z0-9_]{0,31}\}$`)
)

func CanonicalRoute(raw string) (string, error) {
	route := strings.TrimSpace(raw)
	if index := strings.IndexAny(route, "?#"); index >= 0 {
		route = route[:index]
	}
	decoded, err := url.PathUnescape(route)
	if err != nil || decoded == "" || decoded[0] != '/' || strings.Contains(decoded, "//") {
		return "", domain.ErrInvalid
	}
	segments := strings.Split(decoded, "/")
	for index := 1; index < len(segments); index++ {
		segment := segments[index]
		if segment == "" && index == len(segments)-1 {
			continue
		}
		if instanceSegment(segment) {
			segments[index] = "{id}"
		}
	}
	canonical := strings.Join(segments, "/")
	if !domain.ValidRoutePattern(canonical) {
		return "", domain.ErrInvalid
	}
	return canonical, nil
}

func instanceSegment(value string) bool {
	if placeholder.MatchString(value) {
		return false
	}
	if parsed, err := uuid.Parse(value); err == nil && parsed != uuid.Nil {
		return true
	}
	return decimalSegment.MatchString(value) || hexSegment.MatchString(value) ||
		ulidSegment.MatchString(value) || opaqueSegment.MatchString(value)
}
