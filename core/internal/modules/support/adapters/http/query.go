package supporthttp

import (
	"net/http"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func contactFilter(request *http.Request) (domain.ContactFilter, error) {
	query := request.URL.Query()
	if !onlyQueryKeys(query, "page", "page_size", "status", "contact_type") {
		return domain.ContactFilter{}, applicationInvalid()
	}
	page, err := legacyPage(query.Get("page"), query.Get("page_size"), 20)
	if err != nil {
		return domain.ContactFilter{}, err
	}
	filter := domain.ContactFilter{Page: page, Status: domain.ContactStatus(query.Get("status")),
		ContactType: query.Get("contact_type")}
	return filter, nil
}

func bugFilter(request *http.Request, defaultSize int, allowTenant bool) (domain.BugFilter, error) {
	query := request.URL.Query()
	allowed := []string{"page", "page_size", "status", "severity", "category"}
	if allowTenant {
		allowed = append(allowed, "tenant_id")
	}
	if !onlyQueryKeys(query, allowed...) {
		return domain.BugFilter{}, applicationInvalid()
	}
	page, err := legacyPage(query.Get("page"), query.Get("page_size"), defaultSize)
	if err != nil {
		return domain.BugFilter{}, err
	}
	filter := domain.BugFilter{Page: page, Status: domain.BugStatus(query.Get("status")),
		Severity: domain.BugSeverity(query.Get("severity")),
		Category: domain.BugCategory(query.Get("category"))}
	if allowTenant && query.Get("tenant_id") != "" {
		var tenantID uuid.UUID
		tenantID, err = exactUUID(query.Get("tenant_id"))
		filter.TenantID = domain.ID(tenantID)
	}
	return filter, err
}

func legacyPage(pageValue, sizeValue string, defaultSize int) (domain.Page, error) {
	page, size := 1, defaultSize
	var err error
	if pageValue != "" {
		page, err = positiveDecimal(pageValue)
		if err != nil {
			return domain.Page{}, err
		}
	}
	if sizeValue != "" {
		size, err = positiveDecimal(sizeValue)
		if err != nil {
			return domain.Page{}, err
		}
	}
	result := domain.Page{Number: page, Size: size}
	if !result.Valid() {
		return domain.Page{}, applicationInvalid()
	}
	return result, nil
}

func positiveDecimal(value string) (int, error) {
	if value == "" || value[0] == '0' || len(value) > 6 {
		return 0, applicationInvalid()
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, applicationInvalid()
	}
	return parsed, nil
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

func exactUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil || id.String() != value {
		return uuid.Nil, applicationInvalid()
	}
	return id, nil
}

func pathUUID(request *http.Request) (uuid.UUID, error) {
	if request == nil {
		return uuid.Nil, applicationInvalid()
	}
	return exactUUID(request.PathValue("id"))
}
