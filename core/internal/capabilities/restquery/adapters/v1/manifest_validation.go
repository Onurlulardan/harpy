package v1query

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/google/uuid"
)

const (
	maximumLegacyLimit = 1000
	maximumAliasCount  = 256
)

func newManifest(snapshot ManifestSnapshot) (manifest, error) {
	direction, directionOK := api.ParseDirection(snapshot.DefaultSortDirection)
	defaultSort, sortErr := api.NewRequestSort(snapshot.DefaultSortField, direction)
	identity, identityErr := contractretirement.NewIdentity(
		snapshot.OwnerModule, snapshot.Domain,
	)
	if !validManifestID(snapshot.ManifestID) || !directionOK || sortErr != nil ||
		identityErr != nil ||
		snapshot.DefaultLimit < 1 || snapshot.DefaultLimit > maximumLegacyLimit ||
		snapshot.MaximumPageSize < 1 || snapshot.MaximumPageSize > 500 ||
		snapshot.FilterFields == nil || snapshot.SortFields == nil ||
		snapshot.OperatorAliases == nil || snapshot.QueryKeyAliases == nil {
		return manifest{}, fmt.Errorf("verified V1 query manifest is invalid")
	}
	result := manifest{
		id: snapshot.ManifestID, identity: identity,
		defaultLimit:    snapshot.DefaultLimit,
		maximumPageSize: snapshot.MaximumPageSize, defaultSort: defaultSort,
		validated: true,
	}
	var err error
	result.filterFields, err = validateFilterFieldBindings(snapshot.FilterFields)
	if err != nil {
		return manifest{}, err
	}
	result.sortFields, err = validateSortFieldBindings(snapshot.SortFields)
	if err != nil {
		return manifest{}, err
	}
	result.operatorAliases, err = validateOperatorBindings(snapshot.OperatorAliases)
	if err != nil {
		return manifest{}, err
	}
	result.queryKeyAliases, err = validateQueryKeyBindings(snapshot.QueryKeyAliases)
	if err != nil {
		return manifest{}, err
	}
	if err := validateGlobalAliasIDs(result); err != nil {
		return manifest{}, err
	}
	return result, nil
}

func (policy manifest) Validate() error {
	_, sortErr := api.NewRequestSort(
		policy.defaultSort.Field(), policy.defaultSort.Direction(),
	)
	if !policy.validated || !validManifestID(policy.id) ||
		policy.identity.Validate() != nil || sortErr != nil ||
		policy.defaultLimit < 1 || policy.defaultLimit > maximumLegacyLimit ||
		policy.maximumPageSize < 1 || policy.maximumPageSize > 500 ||
		policy.filterFields == nil || policy.sortFields == nil ||
		policy.operatorAliases == nil || policy.queryKeyAliases == nil {
		return fmt.Errorf("verified V1 query manifest is invalid")
	}
	return nil
}

func validateGlobalAliasIDs(policy manifest) error {
	seen := make(map[string]struct{})
	add := func(aliasID string) bool {
		if aliasID == "" {
			return true
		}
		if _, exists := seen[aliasID]; exists {
			return false
		}
		seen[aliasID] = struct{}{}
		return len(seen) <= maximumAliasCount
	}
	for _, binding := range policy.filterFields {
		if !add(binding.aliasID) {
			return fmt.Errorf("verified V1 alias identities are invalid")
		}
	}
	for _, binding := range policy.sortFields {
		if !add(binding.aliasID) {
			return fmt.Errorf("verified V1 alias identities are invalid")
		}
	}
	for _, binding := range policy.operatorAliases {
		if !add(binding.aliasID) {
			return fmt.Errorf("verified V1 alias identities are invalid")
		}
	}
	for _, binding := range policy.queryKeyAliases {
		if !add(binding.aliasID) {
			return fmt.Errorf("verified V1 alias identities are invalid")
		}
	}
	return nil
}

func validManifestID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value &&
		parsed.Variant() == uuid.RFC4122 &&
		parsed.Version() >= 1 && parsed.Version() <= 8
}

func validateOperatorBindings(
	inputs map[string]OperatorBinding,
) (map[string]operatorBinding, error) {
	result := make(map[string]operatorBinding, len(inputs))
	for raw, input := range inputs {
		operator, valid := api.ParseOperator(input.Operator)
		_, builtIn := legacyOperator(raw)
		if !validWireName(raw) || !valid || builtIn || !validAliasID(input.AliasID) {
			return nil, fmt.Errorf("verified V1 operator binding is invalid")
		}
		result[raw] = operatorBinding{operator: operator, aliasID: input.AliasID}
	}
	return result, nil
}

func validateQueryKeyBindings(
	inputs map[string]QueryKeyBinding,
) (map[string]queryKeyBinding, error) {
	result := make(map[string]queryKeyBinding, len(inputs))
	for raw, input := range inputs {
		if !validWireName(raw) || canonicalQueryKey(raw) ||
			!canonicalSimpleKey(input.Key) || !validAliasID(input.AliasID) {
			return nil, fmt.Errorf("verified V1 query-key binding is invalid")
		}
		result[raw] = queryKeyBinding{key: input.Key, aliasID: input.AliasID}
	}
	return result, nil
}

func validOptionalAlias(required bool, value string) bool {
	if required {
		return validAliasID(value)
	}
	return value == ""
}

func validAliasID(value string) bool {
	if len(value) < 1 || len(value) > 128 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if character < 'a' || character > 'z' {
			if character < '0' || character > '9' {
				if !strings.ContainsRune("._:-", character) {
					return false
				}
			}
		}
	}
	return true
}

func validWireName(value string) bool {
	return value != "" && utf8.ValidString(value) && !strings.ContainsRune(value, '\x00')
}
