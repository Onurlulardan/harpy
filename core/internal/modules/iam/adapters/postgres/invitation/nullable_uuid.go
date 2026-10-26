package postgres

import "github.com/google/uuid"

func nullableInvitationUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}
