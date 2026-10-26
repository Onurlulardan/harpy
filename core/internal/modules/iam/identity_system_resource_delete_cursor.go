package iam

import (
	"bytes"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func identityUserDeleteAssignmentCursor(
	after identityfacade.IdentityID,
	rows []identityfacade.Assignment,
) (identityfacade.IdentityID, bool, error) {
	if len(rows) > identityUserDeleteRelationBatchSize {
		return identityfacade.IdentityID{}, false, ErrIdentityHTTPUnavailable
	}
	if len(rows) == 0 {
		return identityfacade.IdentityID{}, true, nil
	}
	next := rows[len(rows)-1].ID
	if !identityUserDeleteCursorAdvances(after, next) {
		return identityfacade.IdentityID{}, false, ErrIdentityHTTPUnavailable
	}
	return next, len(rows) < identityUserDeleteRelationBatchSize, nil
}

func identityUserDeleteMembershipCursor(
	after identityfacade.IdentityID,
	rows []identityfacade.Membership,
) (identityfacade.IdentityID, bool, error) {
	if len(rows) > identityUserDeleteRelationBatchSize {
		return identityfacade.IdentityID{}, false, ErrIdentityHTTPUnavailable
	}
	if len(rows) == 0 {
		return identityfacade.IdentityID{}, true, nil
	}
	next := rows[len(rows)-1].ID
	if !identityUserDeleteCursorAdvances(after, next) {
		return identityfacade.IdentityID{}, false, ErrIdentityHTTPUnavailable
	}
	return next, len(rows) < identityUserDeleteRelationBatchSize, nil
}

func identityUserDeleteCursorAdvances(
	after identityfacade.IdentityID,
	next identityfacade.IdentityID,
) bool {
	if next == (identityfacade.IdentityID{}) {
		return false
	}
	return after == (identityfacade.IdentityID{}) ||
		bytes.Compare(next[:], after[:]) > 0
}
