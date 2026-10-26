package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) LockGlobalBillingMembers(
	ctx context.Context,
	environmentID uuid.UUID,
) ([]string, error) {
	return transaction.readGlobalBillingMembers(
		ctx, lockGlobalBillingMembersSQL, environmentID,
	)
}

func (transaction *RuntimeTransaction) VerifyGlobalBillingMembers(
	ctx context.Context,
	environmentID uuid.UUID,
) ([]string, error) {
	return transaction.readGlobalBillingMembers(
		ctx, verifyGlobalBillingMembersSQL, environmentID,
	)
}

func (transaction *RuntimeTransaction) readGlobalBillingMembers(
	ctx context.Context,
	document string,
	environmentID uuid.UUID,
) ([]string, error) {
	if !transaction.usable() || environmentID == uuid.Nil {
		return nil, ErrAuthorityMismatch
	}
	rows, err := transaction.transaction.Query(ctx, document, environmentID)
	if err != nil {
		return nil, ErrAuthorityMismatch
	}
	defer rows.Close()
	members := make([]string, 0, 1)
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
