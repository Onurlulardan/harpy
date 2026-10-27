package v1query

import (
	"fmt"
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type pendingResult struct {
	request    api.QueryRequest
	page       PagePlan
	manifestID string
	aliasUses  []aliasUse
	validated  bool
}

func newPendingResult(
	request api.QueryRequest, page PagePlan, manifestID string, aliasUses []aliasUse,
) (pendingResult, error) {
	owned := append([]aliasUse{}, aliasUses...)
	slices.SortFunc(owned, compareAliasUses)
	owned = slices.CompactFunc(owned, func(left, right aliasUse) bool {
		return compareAliasUses(left, right) == 0
	})
	result := pendingResult{
		request: request, page: page, manifestID: manifestID,
		aliasUses: owned, validated: true,
	}
	if err := result.Validate(); err != nil {
		return pendingResult{}, err
	}
	return result, nil
}

func (result pendingResult) Request() (api.QueryRequest, error) {
	if err := result.Validate(); err != nil {
		return api.QueryRequest{}, err
	}
	return result.request, nil
}

func (result pendingResult) PagePlan() (PagePlan, error) {
	if err := result.Validate(); err != nil {
		return PagePlan{}, err
	}
	return result.page, nil
}

func (result pendingResult) ManifestID() (string, error) {
	if err := result.Validate(); err != nil {
		return "", err
	}
	return result.manifestID, nil
}

func (result pendingResult) AliasIDs() ([]string, error) {
	if err := result.Validate(); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(result.aliasUses))
	for _, used := range result.aliasUses {
		ids = append(ids, used.aliasID)
	}
	return slices.Compact(ids), nil
}

func (result pendingResult) Validate() error {
	if !result.validated || result.request.Validate() != nil ||
		result.page.Validate() != nil || !validManifestID(result.manifestID) ||
		result.aliasUses == nil || !slices.IsSortedFunc(result.aliasUses, compareAliasUses) {
		return fmt.Errorf("V1 REST Query adapter result is invalid")
	}
	requestPage := result.request.Page()
	_, hasAfter := requestPage.After()
	if result.request.Mode() != api.QueryModeRecord || hasAfter ||
		requestPage.Size() != result.page.limit ||
		result.request.Count() != api.CountModeExact ||
		len(result.request.Select()) != 0 || len(result.request.Includes()) != 0 ||
		len(result.request.Sort()) != 1 {
		return fmt.Errorf("V1 REST Query adapter result is invalid")
	}
	for index, used := range result.aliasUses {
		if !validAliasID(used.aliasID) ||
			index > 0 && compareAliasUses(result.aliasUses[index-1], used) == 0 {
			return fmt.Errorf("V1 REST Query adapter result is invalid")
		}
		if _, err := api.NewQueryIssue(used.path, used.issue); err != nil {
			return fmt.Errorf("V1 REST Query adapter result is invalid")
		}
	}
	return nil
}

func (pendingResult) String() string                        { return v1Redaction }
func (pendingResult) GoString() string                      { return v1Redaction }
func (result pendingResult) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (pendingResult) MarshalJSON() ([]byte, error)          { return rejectV1JSON() }
