package profilebackfill

import (
	"context"
	"crypto/sha256"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func (value *store) Read(
	ctx context.Context,
	scope tenantScope,
	query profilebackfillapp.ReadQuery,
) ([]profilebackfillapp.SourceResource, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope) {
		return nil, profilebackfillapp.ErrInvalid
	}
	if scope.ActorID() != query.Evidence.ActorID {
		return nil, profilebackfillapp.ErrInvalid
	}
	bound, err := value.bind(scope)
	if err != nil {
		return nil, err
	}
	rows, err := bound.database.Query(
		ctx, readSQL, query.ResourceKind, query.SourceSchemaID,
		nullableResourceID(query.AfterResourceID), query.Limit,
		query.Evidence.ActorID, query.Evidence.ManagePermissionID,
		query.Evidence.ExpectedTokenVersion,
		query.Evidence.ExpectedAuthorizationVersion,
	)
	if err != nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	defer rows.Close()
	result := make([]profilebackfillapp.SourceResource, 0, query.Limit)
	for rows.Next() {
		var resource profilebackfillapp.SourceResource
		var digest []byte
		if err := rows.Scan(
			&resource.ID, &resource.SchemaVersionID, &resource.Revision,
			&resource.Document, &digest,
		); err != nil || len(digest) != sha256.Size {
			return nil, profilebackfillapp.ErrUnavailable
		}
		copy(resource.SHA256[:], digest)
		clear(digest)
		if resource.Validate() != nil || len(result) == query.Limit {
			return nil, profilebackfillapp.ErrUnavailable
		}
		result = append(result, resource)
	}
	if rows.Err() != nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return result, nil
}

func (value *store) Lookup(
	ctx context.Context,
	scope tenantScope,
	evidence profilebackfillapp.Evidence,
) (profilebackfillapp.Receipt, bool, error) {
	if evidence.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != evidence.ActorID {
		return profilebackfillapp.Receipt{}, false, profilebackfillapp.ErrInvalid
	}
	bound, err := value.bind(scope)
	if err != nil {
		return profilebackfillapp.Receipt{}, false, err
	}
	receipt, found, err := scanReceipt(bound.database.QueryRow(
		ctx, readReceiptSQL, evidence.OperationID, evidence.ActorID,
		evidence.ManagePermissionID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	))
	if err != nil || !found {
		return profilebackfillapp.Receipt{}, found, err
	}
	return receipt, true, nil
}

type receiptRow interface {
	Scan(...any) error
}

func scanReceipt(row receiptRow) (profilebackfillapp.Receipt, bool, error) {
	var receipt profilebackfillapp.Receipt
	var request []byte
	var target string
	err := row.Scan(
		&receipt.CommandKey, &request, &receipt.ResultID, &target,
		&receipt.Revision, &receipt.AuthorizationClock,
	)
	if receiptMissing(err) {
		return profilebackfillapp.Receipt{}, false, nil
	}
	if err != nil || len(request) != sha256.Size {
		return profilebackfillapp.Receipt{}, false, profilebackfillapp.ErrUnavailable
	}
	copy(receipt.RequestSHA256[:], request)
	clear(request)
	if target != "" {
		receipt.TargetSchemaID, err = uuid.Parse(target)
	}
	if err != nil || receipt.Validate() != nil {
		return profilebackfillapp.Receipt{}, false, profilebackfillapp.ErrUnavailable
	}
	return receipt, true, nil
}

func nullableResourceID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
