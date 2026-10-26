package providerapp

import "github.com/google/uuid"

type SubmitCommandRequest struct {
	ProviderObjectMappingID *uuid.UUID
	SubscriptionID          uuid.UUID
	Kind                    CommandKind
	EffectKey               string
	RequestSHA256           [32]byte
}

func (request SubmitCommandRequest) valid() bool {
	return request.SubscriptionID != uuid.Nil &&
		optionalCommandID(request.ProviderObjectMappingID) && request.Kind.Valid() &&
		validCommandText(request.EffectKey, 256) && request.RequestSHA256 != [32]byte{}
}

func optionalCommandID(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}
