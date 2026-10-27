package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) ListCutoverGroupMembers(
	ctx context.Context,
	tenantID uuid.UUID,
	groupID string,
) ([]string, error) {
	if !transaction.usable() || !schemaNamePattern.MatchString(groupID) {
		return nil, ErrAuthorityMismatch
	}
	rows, err := transaction.transaction.Query(ctx, listCutoverGroupMembersSQL, tenantID, groupID)
	if err != nil {
		return nil, ErrAuthorityMismatch
	}
	defer rows.Close()
	members := make([]string, 0, 2)
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, ErrAuthorityMismatch
		}
		members = append(members, domain)
	}
	if rows.Err() != nil {
		return nil, ErrAuthorityMismatch
	}
	return members, nil
}
