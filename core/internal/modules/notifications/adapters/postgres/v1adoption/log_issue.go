package v1adoption

import (
	"context"
	"crypto/sha256"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	"github.com/google/uuid"
)

var issueNamespace = uuid.MustParse("e89d7d03-f13e-4f07-a849-5f978f8ed3de")

func (store *Store) LogIssue(ctx context.Context, runID, sourceID uuid.UUID,
	sourceDigest [32]byte, code string,
) error {
	if !store.usable(ctx) || runID == uuid.Nil || sourceID == uuid.Nil ||
		sourceDigest == [32]byte{} || code == "" {
		return app.ErrInvalid
	}
	keyDigest := sha256.Sum256([]byte(sourceID.String()))
	identity := append(append([]byte(runID.String()+"\x00"+sourceID.String()+"\x00"+code),
		sourceDigest[:]...), keyDigest[:]...)
	issueID := uuid.NewSHA1(issueNamespace, identity)
	if err := store.execute(ctx, insertIssueSQL, issueID, runID,
		keyDigest[:], code); err != nil {
		return app.ErrUnavailable
	}
	return nil
}
