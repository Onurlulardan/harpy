package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) ResolveJobControl(
	ctx context.Context,
	authority JobAuthority,
	jobID uuid.UUID,
) (JobControl, error) {
	if jobID == uuid.Nil {
		return JobControl{}, ErrJobControlInvalid
	}
	return withJobAuthorityTransaction(
		ctx, repository, authority,
		func(transaction pgx.Tx) (JobControl, error) {
			arguments := append(authorityArguments(authority), jobID)
			var found uuid.UUID
			err := transaction.QueryRow(
				ctx, resolveJobControlSQL, arguments...,
			).Scan(&found)
			if errors.Is(err, pgx.ErrNoRows) {
				return JobControl{}, ErrJobNotControllable
			}
			if err != nil {
				return JobControl{}, fmt.Errorf("resolve job control: %w", err)
			}
			return newJobControl(repository.identity, authority, found)
		},
	)
}
