package compatibility

import (
	"context"
)

func (application *Application) BackfillItem(
	ctx context.Context, authority AdoptionAuthority, item AdoptionItem,
) (AdoptionItem, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validAuthority(authority) || !validAdoptionItem(item) {
		return AdoptionItem{}, ErrInvalid
	}
	item.CreatedAt = application.clock.Now()
	if !validTime(item.CreatedAt) {
		return AdoptionItem{}, ErrUnavailable
	}
	result, _, err := application.adoption.BackfillItem(ctx, authority, item)
	if err != nil {
		return AdoptionItem{}, err
	}
	if !sameAdoptionItem(result, item) {
		return AdoptionItem{}, ErrUnavailable
	}
	return result, nil
}

func validAdoptionItem(item AdoptionItem) bool {
	if !validUUID(item.RunID) || !validUUID(item.LegacyEndpointID) ||
		item.SourceSHA256 == [32]byte{} || !item.Disposition.Valid() {
		return false
	}
	if item.Disposition == DispositionAdopted {
		return item.ReasonCode == "" && validV7(item.QueryID) &&
			validV7(item.QueryVersionID) && item.TargetSHA256 != [32]byte{} &&
			(item.PublicationID == uuidNil || validV7(item.PublicationID))
	}
	return reasonPattern.MatchString(item.ReasonCode) && item.QueryID == uuidNil &&
		item.QueryVersionID == uuidNil && item.PublicationID == uuidNil &&
		item.TargetSHA256 == [32]byte{}
}

var uuidNil = [16]byte{}

func sameAdoptionItem(left, right AdoptionItem) bool {
	return left.RunID == right.RunID && left.LegacyEndpointID == right.LegacyEndpointID &&
		left.SourceSHA256 == right.SourceSHA256 && left.TargetSHA256 == right.TargetSHA256 &&
		left.Disposition == right.Disposition && left.ReasonCode == right.ReasonCode &&
		left.QueryID == right.QueryID && left.QueryVersionID == right.QueryVersionID &&
		left.PublicationID == right.PublicationID && left.CreatedAt.Equal(right.CreatedAt)
}
