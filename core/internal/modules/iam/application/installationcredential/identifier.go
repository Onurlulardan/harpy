package installationcredential

import "github.com/google/uuid"

var credentialNamespace = uuid.MustParse("a89d7c1c-3f57-5a34-b77e-122f5bec19b4")

func InitialCredentialID(operationID, userID uuid.UUID) (uuid.UUID, error) {
	if operationID == uuid.Nil || userID == uuid.Nil {
		return uuid.Nil, ErrInvalid
	}
	name := append(append(make([]byte, 0, 32), operationID[:]...), userID[:]...)
	identifier := uuid.NewSHA1(credentialNamespace, name)
	clear(name)
	if identifier == uuid.Nil || identifier.Version() != uuid.Version(5) {
		return uuid.Nil, ErrInvalid
	}
	return identifier, nil
}
