package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (repository *workerRepository) ListJobAuthorities(
	ctx context.Context,
	token JobAuthorityPageToken,
) (JobAuthorityPage, error) {
	if ctx == nil {
		return JobAuthorityPage{}, ErrWorkerContextInvalid
	}
	if !token.valid(repository) {
		return JobAuthorityPage{}, ErrJobAuthorityToken
	}
	cursor := token.cursor
	rows, err := repository.database.Query(
		ctx,
		listJobAuthoritiesSQL,
		cursor.present,
		cursor.scopeKind,
		cursor.scopeID,
		cursor.groupID,
		cursor.domain,
		token.limit+1,
	)
	if err != nil {
		return JobAuthorityPage{}, fmt.Errorf("list current job authorities: %w", err)
	}
	defer rows.Close()
	authorities := make([]JobAuthority, 0, token.limit+1)
	for rows.Next() {
		var scopeKind, groupID, domain, mode string
		var scopeID uuid.UUID
		var groupGeneration, generation int64
		if err := rows.Scan(
			&scopeKind,
			&scopeID,
			&groupID,
			&groupGeneration,
			&domain,
			&generation,
			&mode,
		); err != nil {
			return JobAuthorityPage{}, fmt.Errorf("scan current job authority: %w", err)
		}
		authority, err := newJobAuthorityWithMode(
			repository.identity,
			jobAuthorityMode(mode),
			scopeKind,
			scopeID,
			groupID,
			groupGeneration,
			domain,
			generation,
		)
		if err != nil {
			return JobAuthorityPage{}, fmt.Errorf("decode current job authority: %w", err)
		}
		authorities = append(authorities, authority)
		if len(authorities) > token.limit+1 {
			return JobAuthorityPage{}, ErrJobAuthorityPageInvalid
		}
	}
	if err := rows.Err(); err != nil {
		return JobAuthorityPage{}, fmt.Errorf("iterate current job authorities: %w", err)
	}
	return makeJobAuthorityPage(token, authorities), nil
}
