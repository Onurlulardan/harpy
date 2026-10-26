package lifecycle

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CredentialLookup struct {
	Purpose       CredentialPurpose
	Selector      [16]byte
	BindingSHA256 [32]byte
}

func (lookup CredentialLookup) Validate() error {
	if !lookup.Purpose.Valid() || lookup.Selector == [16]byte{} ||
		lookup.BindingSHA256 == [32]byte{} {
		return ErrInvalid
	}
	return nil
}

type CredentialRecord struct {
	TokenID, UserID, IdentifierID, PortalBindingID uuid.UUID
	Purpose, IdentifierKind, State                 string
	IdentifierRevision, PortalRevision             int64
	Selector                                       [16]byte
	HashKeyID                                      string
	SecretHMAC, BindingSHA256                      [32]byte
	Revision                                       int64
	ExpiresAt                                      time.Time
}

func (record CredentialRecord) Validate() error {
	purpose := CredentialPurpose(record.Purpose)
	if record.TokenID == uuid.Nil || record.UserID == uuid.Nil ||
		record.IdentifierID == uuid.Nil || !purpose.Valid() ||
		(record.IdentifierKind != "email_v1" && record.IdentifierKind != "portal_v1") ||
		record.IdentifierRevision < 1 || record.Selector == [16]byte{} ||
		record.HashKeyID == "" || len(record.HashKeyID) > 128 ||
		record.SecretHMAC == [32]byte{} || record.BindingSHA256 == [32]byte{} ||
		record.Revision < 1 || (record.State != "pending" && record.State != "consumed") ||
		record.ExpiresAt.IsZero() {
		return ErrInvalid
	}
	if purpose == PurposePortalSetup {
		if record.IdentifierKind != "portal_v1" || record.PortalBindingID == uuid.Nil ||
			record.PortalRevision < 1 {
			return ErrInvalid
		}
	} else if record.PortalBindingID != uuid.Nil || record.PortalRevision != 0 {
		return ErrInvalid
	}
	if purpose == PurposeEmailVerification && record.IdentifierKind != "email_v1" {
		return ErrInvalid
	}
	return nil
}

func (service *Service) LookupCredential(ctx context.Context,
	scope platformtx.CredentialTenantScope, lookup CredentialLookup,
) (CredentialRecord, error) {
	purpose, valid := lookup.Purpose.transactionPurpose()
	if !service.Valid() || lookup.Validate() != nil || !valid ||
		!credentialLookupUsable(ctx, scope, purpose) ||
		scope.LocatorBindingSHA256() != lookup.BindingSHA256 {
		return CredentialRecord{}, ErrInvalid
	}
	record, err := service.store.LookupCredential(ctx, scope, lookup)
	if err != nil {
		return CredentialRecord{}, err
	}
	if record.Validate() != nil || CredentialPurpose(record.Purpose) != lookup.Purpose ||
		record.Selector != lookup.Selector || record.BindingSHA256 != lookup.BindingSHA256 {
		return CredentialRecord{}, ErrUnavailable
	}
	return record, nil
}

func (purpose CredentialPurpose) transactionPurpose() (platformtx.CredentialPurpose, bool) {
	switch purpose {
	case PurposeEmailVerification:
		return platformtx.CredentialEmailVerification, true
	case PurposePasswordReset:
		return platformtx.CredentialPasswordReset, true
	case PurposeInitialAdminSetup:
		return platformtx.CredentialInitialAdminSetup, true
	case PurposePortalSetup:
		return platformtx.CredentialPortalSetup, true
	default:
		return "", false
	}
}
