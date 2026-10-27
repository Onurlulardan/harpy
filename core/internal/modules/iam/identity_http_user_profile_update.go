package iam

import (
	"context"
	"encoding/json"
	"strings"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) userProfileUpdateValues(
	ctx context.Context,
	scope platformtx.TenantScope,
	user identityfacade.User,
	raw json.RawMessage,
	firstName *string,
	lastName *string,
	phone *string,
) ([]byte, error) {
	input, err := identityProfileInputMap(raw)
	if err != nil {
		return nil, err
	}
	aliases, err := application.identities.profiles.FieldIDsForSchema(
		ctx, scope, identityfacade.ProfileUser, user.Profile.SchemaID(),
		identityUserProfileAliasKeys(input),
	)
	if err != nil {
		return nil, err
	}
	values, err := identityApplyProfileAliases(input, aliases)
	if err != nil {
		return nil, err
	}
	for _, key := range identityUserCoreProfileKeys {
		if aliases[key] == (identityfacade.IdentityID{}) {
			return nil, ErrIdentityHTTPUnavailable
		}
	}
	current, err := identityProfileMap(user.Profile.Values())
	if err != nil {
		return nil, ErrIdentityHTTPUnavailable
	}
	effectiveFirst, err := identityUserProfileName(
		current, aliases["first_name"], firstName,
	)
	if err != nil {
		return nil, err
	}
	effectiveLast, err := identityUserProfileName(
		current, aliases["last_name"], lastName,
	)
	if err != nil {
		return nil, err
	}
	for key, value := range map[string]*string{
		"first_name": firstName, "last_name": lastName, "phone": phone,
	} {
		if value != nil {
			if err := identitySetProfileString(values, aliases[key], *value); err != nil {
				return nil, err
			}
		}
	}
	if firstName != nil || lastName != nil {
		display := strings.TrimSpace(effectiveFirst + " " + effectiveLast)
		if display == "" {
			return nil, ErrIdentityHTTPInvalid
		}
		if err := identitySetProfileString(
			values, aliases["display_name"], display,
		); err != nil {
			return nil, err
		}
	}
	return identityMarshalProfile(values)
}

func identityUserProfileAliasKeys(
	input map[string]json.RawMessage,
) []string {
	keys := append([]string(nil), identityUserCoreProfileKeys...)
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		seen[key] = struct{}{}
	}
	for key := range input {
		_, duplicate := seen[key]
		if _, err := uuid.Parse(key); err != nil && !duplicate {
			keys = append(keys, key)
			seen[key] = struct{}{}
		}
	}
	return keys
}

func identityApplyProfileAliases(
	input map[string]json.RawMessage,
	aliases map[string]identityfacade.IdentityID,
) (map[string]json.RawMessage, error) {
	values := make(map[string]json.RawMessage, len(input))
	for key, value := range input {
		target := key
		if id, found := aliases[key]; found {
			target = uuid.UUID(id).String()
		} else if _, err := uuid.Parse(key); err != nil {
			return nil, ErrIdentityHTTPInvalid
		}
		if _, collision := values[target]; collision {
			return nil, ErrIdentityHTTPInvalid
		}
		values[target] = value
	}
	return values, nil
}

func identityUserProfileName(
	current map[string]json.RawMessage,
	id identityfacade.IdentityID,
	replacement *string,
) (string, error) {
	if replacement != nil {
		return *replacement, nil
	}
	raw, found := current[uuid.UUID(id).String()]
	if !found {
		return "", nil
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", ErrIdentityHTTPUnavailable
	}
	return value, nil
}

func identitySetProfileString(
	values map[string]json.RawMessage,
	id identityfacade.IdentityID,
	value string,
) error {
	key := uuid.UUID(id).String()
	if _, collision := values[key]; collision {
		return ErrIdentityHTTPInvalid
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ErrIdentityHTTPInvalid
	}
	values[key] = encoded
	return nil
}
