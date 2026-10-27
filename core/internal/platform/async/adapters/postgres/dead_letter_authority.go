package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type deadLetterAuthority struct {
	scopeKind       string
	scopeID         uuid.UUID
	groupID         string
	groupGeneration int64
	domain          string
	generation      int64
}

func resolveDeadLetterAuthority(
	ctx context.Context,
	transaction pgx.Tx,
	scopeKind string,
	scopeID uuid.UUID,
	domain string,
) (deadLetterAuthority, error) {
	rows, err := transaction.Query(
		ctx, resolveDeadLetterAuthoritySQL, scopeKind, scopeID, domain,
	)
	if err != nil {
		return deadLetterAuthority{}, fmt.Errorf("lock dead-letter authority: %w", err)
	}
	authorities := make([]deadLetterAuthority, 0, 2)
	for rows.Next() {
		var authority deadLetterAuthority
		if err := rows.Scan(
			&authority.scopeKind, &authority.scopeID,
			&authority.groupID, &authority.groupGeneration,
			&authority.domain, &authority.generation,
		); err != nil {
			return deadLetterAuthority{}, fmt.Errorf("scan dead-letter authority: %w", err)
		}
		authorities = append(authorities, authority)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return deadLetterAuthority{}, fmt.Errorf("read dead-letter authority: %w", err)
	}
	rows.Close()
	if len(authorities) != 1 {
		return deadLetterAuthority{}, ErrDeadLetterAuthority
	}
	authority := authorities[0]
	if _, err := transaction.Exec(
		ctx, assertJobAuthoritySQL,
		authority.scopeKind, authority.scopeID, authority.groupID,
		authority.groupGeneration, authority.domain, authority.generation,
	); err != nil {
		return deadLetterAuthority{}, fmt.Errorf("assert dead-letter authority: %w", err)
	}
	return authority, nil
}
