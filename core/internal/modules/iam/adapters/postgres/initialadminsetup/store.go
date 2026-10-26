package initialadminsetup

import (
	"context"
	"errors"

	setupapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/initialadminsetup"
)

var ErrUnavailable = errors.New("initial admin setup stage store is unavailable")

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row

type Store struct{ queryRow QueryRow }

func NewStore(queryRow QueryRow) (*Store, error) {
	if queryRow == nil {
		return nil, ErrUnavailable
	}
	return &Store{queryRow: queryRow}, nil
}

func (store *Store) Stage(ctx context.Context, stage setupapp.Stage) error {
	if store == nil || store.queryRow == nil || ctx == nil || ctx.Err() != nil ||
		stage.Validate() != nil {
		return setupapp.ErrInvalid
	}
	var replay bool
	e := stage.Envelope
	err := store.queryRow(ctx, stageSQL, stage.OperationID, stage.TenantID,
		stage.UserID, stage.IdentifierID, stage.Email, stage.Token.ID,
		stage.Token.Selector[:], stage.Token.HashKeyID, stage.Token.SecretHMAC[:],
		stage.Token.BindingSHA256[:], e.ArtifactID, stage.Reference, e.KeyID,
		e.Nonce[:], e.Ciphertext[:], e.Tag[:], e.AADSHA256[:],
		e.CiphertextSHA256[:], stage.CreatedAt).Scan(&replay)
	if err != nil {
		return err
	}
	return nil
}

var _ setupapp.StageStore = (*Store)(nil)
