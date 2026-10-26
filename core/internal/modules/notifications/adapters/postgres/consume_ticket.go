package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) InsertTicket(ctx context.Context, value domain.Ticket) error {
	if !store.usable(ctx, true) || !validTicket(value) || value.ConsumedAt != nil {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertRealtimeTicketSQL, value.ID.String(),
		value.TenantID.String(), value.UserID.String(), value.TokenHash[:],
		value.AuthorizationVersion, value.Origin, value.CreatedAt, value.ExpiresAt)
	return err
}

func (store *Store) ConsumeTicket(ctx context.Context,
	input domain.TicketConsume,
) (domain.Ticket, error) {
	if !store.usable(ctx, true) || !input.TenantID.Valid() || !input.UserID.Valid() ||
		input.AuthorizationVersion < 1 || input.Origin == "" || input.Token == "" ||
		input.At.IsZero() {
		return domain.Ticket{}, ErrInvalid
	}
	hash := sha256.Sum256([]byte(input.Token))
	value, err := scanTicket(store.queryRow(ctx, consumeRealtimeTicketSQL, hash[:],
		input.TenantID.String(), input.UserID.String(), input.AuthorizationVersion,
		input.Origin, input.At.UTC()))
	if store.notFound(err) {
		return domain.Ticket{}, domain.ErrTicketConflict
	}
	return value, err
}

func scanTicket(row scanner) (domain.Ticket, error) {
	var value domain.Ticket
	var id, tenant, user string
	var hash []byte
	err := row.Scan(&id, &tenant, &user, &hash, &value.AuthorizationVersion,
		&value.Origin, &value.CreatedAt, &value.ExpiresAt, &value.ConsumedAt)
	if err != nil {
		return domain.Ticket{}, err
	}
	value.ID, err = decodeID(id)
	if err == nil {
		value.TenantID, err = decodeID(tenant)
	}
	if err == nil {
		value.UserID, err = decodeID(user)
	}
	if err == nil {
		value.TokenHash, err = decodeDigest(hash)
	}
	if err != nil || !validTicket(value) {
		return domain.Ticket{}, fmt.Errorf("decode notification realtime ticket")
	}
	return value, nil
}

func validTicket(value domain.Ticket) bool {
	return value.ID.Valid() && value.TenantID.Valid() && value.UserID.Valid() &&
		value.TokenHash != [32]byte{} && value.AuthorizationVersion > 0 &&
		value.Origin != "" && !value.CreatedAt.IsZero() &&
		value.ExpiresAt.Sub(value.CreatedAt) == domain.TicketLifetime
}
