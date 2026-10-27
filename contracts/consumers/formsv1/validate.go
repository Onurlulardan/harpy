package formsv1

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

func (contract Contract) Validate() error {
	if contract.Version != 1 || contract.ContractID != "forms-record-list-consumers.v1" ||
		!exactSorted(contract.Consumers, []string{"frontend-v3"}) ||
		len(contract.Routes) != 2 || !sortedUnique(contract.QueryKeys) ||
		len(contract.QueryKeys) == 0 || len(contract.SuccessPointers) == 0 {
		return fmt.Errorf("Forms V1 consumer contract identity is invalid")
	}
	seenRoutes := make(map[string]bool, len(contract.Routes))
	for _, route := range contract.Routes {
		if err := validateRoute(route, seenRoutes); err != nil {
			return err
		}
	}
	if !sortedUnique(contract.SuccessPointers) {
		return fmt.Errorf("Forms V1 success pointers are invalid")
	}
	return nil
}

func validateRoute(route Route, seen map[string]bool) error {
	if route.ID == "" || seen[route.ID] || route.Method != "GET" ||
		!strings.HasPrefix(route.Path, "/api/") ||
		!strings.Contains(route.Path, "{slug}") || route.OperationID == "" ||
		route.SuccessStatus != 0 && route.SuccessStatus != 200 || len(route.Errors) == 0 {
		return fmt.Errorf("Forms V1 consumer route %q is invalid", route.ID)
	}
	seen[route.ID] = true
	seenErrors := make(map[string]bool, len(route.Errors))
	for _, contract := range route.Errors {
		if contract.Scenario == "" || seenErrors[contract.Scenario] ||
			contract.Status < 400 || contract.Status > 599 || !stableCode(contract.Code) {
			return fmt.Errorf("Forms V1 consumer route %q error is invalid", route.ID)
		}
		seenErrors[contract.Scenario] = true
	}
	return nil
}

func stableCode(value string) bool {
	if value == "" {
		return false
	}
	for _, current := range value {
		if current != '_' && !unicode.IsUpper(current) && !unicode.IsDigit(current) {
			return false
		}
	}
	return true
}

func sortedUnique(values []string) bool {
	return len(values) != 0 && slices.IsSorted(values) &&
		len(slices.Compact(append([]string{}, values...))) == len(values)
}

func exactSorted(actual, expected []string) bool {
	return slices.Equal(actual, expected) && sortedUnique(actual)
}
