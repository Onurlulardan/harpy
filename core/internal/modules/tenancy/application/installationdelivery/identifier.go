package installationdelivery

import "github.com/google/uuid"

var initialAttemptNamespace = uuid.MustParse("0b983b18-086e-5f78-8a13-251bc8c6b837")

func InitialAttemptID(operationID, credentialID uuid.UUID) (uuid.UUID, error) {
	if operationID == uuid.Nil || credentialID == uuid.Nil {
		return uuid.Nil, ErrInvalid
	}
	name := append(append(make([]byte, 0, 32), operationID[:]...), credentialID[:]...)
	identifier := uuid.NewSHA1(initialAttemptNamespace, name)
	clear(name)
	if identifier == uuid.Nil || identifier.Version() != uuid.Version(5) {
		return uuid.Nil, ErrInvalid
	}
	return identifier, nil
}
