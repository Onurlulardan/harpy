package authentication

import "github.com/google/uuid"

type identifierSet struct {
	session uuid.UUID
	family  uuid.UUID
	token   uuid.UUID
	jti     uuid.UUID
}

type rotationIdentifiers struct {
	token uuid.UUID
	jti   uuid.UUID
}

func (authenticator Authenticator) nextIdentifierSet() (identifierSet, error) {
	values := make([]uuid.UUID, 4)
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index := range values {
		value, err := authenticator.identifiers()
		if err != nil || value == uuid.Nil {
			return identifierSet{}, ErrAuthenticationCrypto
		}
		if _, exists := seen[value]; exists {
			return identifierSet{}, ErrAuthenticationCrypto
		}
		seen[value] = struct{}{}
		values[index] = value
	}
	return identifierSet{
		session: values[0], family: values[1], token: values[2], jti: values[3],
	}, nil
}

func (authenticator Authenticator) nextRotationIdentifiers() (rotationIdentifiers, error) {
	first, err := authenticator.identifiers()
	if err != nil || first == uuid.Nil {
		return rotationIdentifiers{}, ErrAuthenticationCrypto
	}
	second, err := authenticator.identifiers()
	if err != nil || second == uuid.Nil || second == first {
		return rotationIdentifiers{}, ErrAuthenticationCrypto
	}
	return rotationIdentifiers{token: first, jti: second}, nil
}
