package menu

import (
	"math"
	"time"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

func validMutation(value menuapp.OverrideMutation) bool {
	if !value.Override.ItemID.Valid() || value.ReceiptID == uuid.Nil ||
		value.ReceiptID.Version() != 7 || value.ActorID == uuid.Nil ||
		value.ReceiptID == value.ActorID || value.ExpectedRevision > math.MaxInt64 ||
		!idempotencyPattern.MatchString(value.IdempotencyKey) ||
		!digestPattern.MatchString(value.RequestSHA256) || !canonicalTime(value.OccurredAt) ||
		value.Override.Label == nil && value.Override.Order == nil && value.Override.Visible == nil {
		return false
	}
	_, err := menudomain.ApplyOverrides([]menudomain.Item{{
		MenuItem: packagecontract.MenuItem{ID: value.Override.ItemID, Label: "validation"},
		Visible:  true,
	}}, []menudomain.Override{value.Override})
	return err == nil
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
