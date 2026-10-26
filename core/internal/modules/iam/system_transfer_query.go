package iam

import (
	"context"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const systemTransferMaximumQueryBytes = 262144

type systemTransferPreparedQuery struct {
	document []byte
	digest   [32]byte
	catalog  systemTransferCatalogPin
}

type systemTransferSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func (runtime *SystemTransferRuntime) prepareSystemExportQuery(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource SystemTransferResource,
	document []byte,
) (systemTransferPreparedQuery, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, systemTransferMaximumQueryBytes,
	)
	if err != nil {
		return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(canonical, &object); err != nil {
		return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
	}
	for _, forbidden := range []string{"page", "count", "include", "aggregate"} {
		if _, present := object[forbidden]; present {
			return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
		}
	}
	if err = systemTransferStableSort(object); err != nil {
		return systemTransferPreparedQuery{}, err
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
	}
	canonical, err = canonicaljsonv1.CanonicalizeObject(
		encoded, systemTransferMaximumQueryBytes,
	)
	if err != nil {
		return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
	}
	decoded, err := queryapp.DecodePOST(canonical)
	if err != nil {
		return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
	}
	published, err := runtime.queries.ResolveCatalog(ctx, scope, string(resource))
	if err != nil {
		return systemTransferPreparedQuery{}, systemTransferError(err)
	}
	if err = queryapp.ValidateCatalogRequest(published, decoded); err != nil {
		return systemTransferPreparedQuery{}, ErrSystemTransferInvalid
	}
	pin, err := systemTransferPublishedPin(published)
	if err != nil {
		return systemTransferPreparedQuery{}, err
	}
	return systemTransferPreparedQuery{
		document: canonical, digest: systemTransferSHA256(canonical), catalog: pin,
	}, nil
}

func systemTransferStableSort(object map[string]json.RawMessage) error {
	var sorts []systemTransferSort
	if raw, present := object["sort"]; present {
		if err := json.Unmarshal(raw, &sorts); err != nil {
			return ErrSystemTransferInvalid
		}
	}
	if len(sorts) > 4 {
		return ErrSystemTransferInvalid
	}
	for index, sort := range sorts {
		if sort.Field == "id" {
			if index != len(sorts)-1 {
				return ErrSystemTransferInvalid
			}
			object["sort"], _ = json.Marshal(sorts)
			return nil
		}
	}
	sorts = append(sorts, systemTransferSort{Field: "id", Direction: "asc"})
	object["sort"], _ = json.Marshal(sorts)
	return nil
}
