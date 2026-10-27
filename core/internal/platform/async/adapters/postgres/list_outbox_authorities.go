package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (repository *workerRepository) ListOutboxAuthorities(
	ctx context.Context,
	token OutboxAuthorityPageToken,
) (OutboxAuthorityPage, error) {
	if ctx == nil {
		return OutboxAuthorityPage{}, ErrWorkerContextInvalid
	}
	if !token.valid(repository) {
		return OutboxAuthorityPage{}, ErrOutboxAuthorityToken
	}
	cursor := token.cursor
	rows, err := repository.database.Query(
		ctx, listOutboxAuthoritiesSQL, cursor.present, cursor.scopeKind,
		cursor.scopeID, cursor.groupID, cursor.domain, token.limit+1,
	)
	if err != nil {
		return OutboxAuthorityPage{}, fmt.Errorf("list current outbox authorities: %w", err)
	}
	defer rows.Close()
	authorities := make([]OutboxAuthority, 0, token.limit+1)
	for rows.Next() {
		var scopeKind, groupID, domain string
		var scopeID uuid.UUID
		var groupGeneration, generation int64
		if err := rows.Scan(
			&scopeKind, &scopeID, &groupID, &groupGeneration, &domain, &generation,
		); err != nil {
			return OutboxAuthorityPage{}, fmt.Errorf("scan current outbox authority: %w", err)
		}
		authority, err := newOutboxAuthority(
			repository.identity, scopeKind, scopeID, groupID,
			groupGeneration, domain, generation,
		)
		if err != nil {
			return OutboxAuthorityPage{}, fmt.Errorf("decode current outbox authority: %w", err)
		}
		authorities = append(authorities, authority)
		if len(authorities) > token.limit+1 {
			return OutboxAuthorityPage{}, ErrOutboxAuthorityToken
		}
	}
	if err := rows.Err(); err != nil {
		return OutboxAuthorityPage{}, fmt.Errorf("iterate current outbox authorities: %w", err)
	}
	return makeOutboxAuthorityPage(token, authorities), nil
}
