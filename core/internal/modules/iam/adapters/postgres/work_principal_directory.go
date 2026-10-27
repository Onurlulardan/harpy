package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func scanWorkAvailableUsers(
	ctx context.Context,
	database workPrincipalQueryer,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	search string,
	limit int,
) (json.RawMessage, error) {
	var document []byte
	err := database.QueryRow(
		ctx, workAvailableUsersSQL, tenantID, actorID, search, limit,
	).Scan(&document)
	if err != nil || !validWorkAvailableUsers(document) {
		return nil, ErrWorkPrincipalRead
	}
	return append(json.RawMessage(nil), document...), nil
}

func validWorkAvailableUsers(document []byte) bool {
	var users []struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Avatar    string `json:"avatar"`
	}
	return json.Unmarshal(document, &users) == nil && users != nil && len(users) <= 50
}
