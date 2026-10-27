package iam

import (
	"context"
	"encoding/json"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var identityUserCoreProfileKeys = []string{
	"display_name", "first_name", "last_name", "phone",
}

func identityProfileValues(raw json.RawMessage) ([]byte, error) {
	values, err := identityProfileInputMap(raw)
	if err != nil {
		return nil, err
	}
	return identityMarshalProfile(values)
}

func (application *identityHTTPApplication) profileInputValues(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource identityfacade.ProfileResource,
	raw json.RawMessage,
) (map[string]json.RawMessage, error) {
	input, err := identityProfileInputMap(raw)
	if err != nil {
		return nil, err
	}
	publicKeys := make([]string, 0, len(input))
	for key := range input {
		if _, parseErr := uuid.Parse(key); parseErr != nil {
			publicKeys = append(publicKeys, key)
		}
	}
	aliases := map[string]identityfacade.IdentityID{}
	if len(publicKeys) != 0 {
		aliases, err = application.identities.profiles.FieldIDs(
			ctx, scope, resource, publicKeys,
		)
		if err != nil || len(aliases) != len(publicKeys) {
			return nil, ErrIdentityHTTPInvalid
		}
	}
	values := make(map[string]json.RawMessage, len(input))
	for key, value := range input {
		target := key
		if id, found := aliases[key]; found {
			target = uuid.UUID(id).String()
		}
		if _, collision := values[target]; collision {
			return nil, ErrIdentityHTTPInvalid
		}
		values[target] = value
	}
	return values, nil
}

func (application *identityHTTPApplication) profileDocumentValues(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource identityfacade.ProfileResource,
	raw json.RawMessage,
) ([]byte, error) {
	values, err := application.profileInputValues(ctx, scope, resource, raw)
	if err != nil {
		return nil, err
	}
	return identityMarshalProfile(values)
}

func (application *identityHTTPApplication) userProfileValues(
	ctx context.Context,
	scope platformtx.TenantScope,
	raw json.RawMessage,
	firstName *string,
	lastName *string,
	phone *string,
) ([]byte, error) {
	values, err := application.profileInputValues(
		ctx, scope, identityfacade.ProfileUser, raw,
	)
	if err != nil {
		return nil, err
	}
	aliases, err := application.identities.profiles.FieldIDs(
		ctx, scope, identityfacade.ProfileUser, identityUserCoreProfileKeys,
	)
	if err != nil || len(aliases) != len(identityUserCoreProfileKeys) {
		return nil, ErrIdentityHTTPUnavailable
	}
	for key, value := range map[string]*string{
		"first_name": firstName,
		"last_name":  lastName,
		"phone":      phone,
	} {
		if value == nil {
			continue
		}
		id := uuid.UUID(aliases[key]).String()
		if _, collision := values[id]; collision {
			return nil, ErrIdentityHTTPInvalid
		}
		encoded, marshalErr := json.Marshal(*value)
		if marshalErr != nil {
			return nil, ErrIdentityHTTPInvalid
		}
		values[id] = encoded
	}
	if firstName != nil && lastName != nil {
		display, marshalErr := json.Marshal(*firstName + " " + *lastName)
		id := uuid.UUID(aliases["display_name"]).String()
		if marshalErr != nil {
			return nil, ErrIdentityHTTPInvalid
		}
		values[id] = display
	}
	return identityMarshalProfile(values)
}

func identityProfileMap(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(raw) == 0 {
		return map[string]json.RawMessage{}, nil
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil || len(values) > 100 {
		return nil, ErrIdentityHTTPInvalid
	}
	for key, value := range values {
		id, err := uuid.Parse(key)
		if err != nil || id == uuid.Nil || id.String() != key || len(value) == 0 {
			return nil, ErrIdentityHTTPInvalid
		}
	}
	return values, nil
}

func identityProfileInputMap(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(raw) == 0 {
		return map[string]json.RawMessage{}, nil
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil || len(values) > 100 {
		return nil, ErrIdentityHTTPInvalid
	}
	for key, value := range values {
		id, idErr := uuid.Parse(key)
		stableID := idErr == nil && id != uuid.Nil && id.String() == key
		if !stableID && !validIdentityProfileKey(key) || len(value) == 0 {
			return nil, ErrIdentityHTTPInvalid
		}
	}
	return values, nil
}

func identityMarshalProfile(values map[string]json.RawMessage) ([]byte, error) {
	document, err := json.Marshal(values)
	if err != nil || len(document) < 2 || len(document) > identityfacade.MaximumProfileBytes {
		return nil, ErrIdentityHTTPInvalid
	}
	return document, nil
}
