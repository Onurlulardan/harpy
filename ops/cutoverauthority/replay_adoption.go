package main

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

func loadStoredAdoptionReceipt(
	ctx context.Context, query catalogQuerier, bundle verifiedAuthorityBundle,
	signer attestation.Signer,
) (signedAdoptionReceipt, bool, error) {
	operation := bundle.operation.Value()
	var document []byte
	var digest string
	err := query.QueryRow(ctx, `SELECT receipt_document,receipt_sha256
FROM cutover_authority.authority_adoption_inventory
WHERE catalog_generation=$1 AND operation_id=$2`,
		operation.CatalogGeneration, operation.OperationID).Scan(&document, &digest)
	if err != nil {
		if isNoRows(err) {
			return signedAdoptionReceipt{}, false, nil
		}
		return signedAdoptionReceipt{}, false,
			fmt.Errorf("protected authority inventory read failed")
	}
	computed, err := attestation.DigestHex(document)
	if err != nil || computed != digest {
		return signedAdoptionReceipt{}, false,
			fmt.Errorf("stored authority receipt digest differs")
	}
	var receipt cutoverv1.AuthorityAdoptionReceipt
	if err := attestation.DecodeStrict(document, &receipt); err != nil {
		return signedAdoptionReceipt{}, false,
			fmt.Errorf("stored authority receipt decode failed")
	}
	adoptedAt, err := time.Parse(time.RFC3339Nano, receipt.AdoptedAt)
	if err != nil {
		return signedAdoptionReceipt{}, false,
			fmt.Errorf("stored authority receipt timestamp differs")
	}
	signed, err := signAdoptionReceipt(receipt, bundle, signer, adoptedAt)
	if err != nil || signed.digest != digest {
		return signedAdoptionReceipt{}, false,
			fmt.Errorf("stored authority receipt verification failed")
	}
	var ignored any
	if err := query.QueryRow(ctx,
		"SELECT cutover_authority.assert_adopted_catalog_v1($1::text)", digest).Scan(&ignored); err != nil {
		return signedAdoptionReceipt{}, false,
			fmt.Errorf("stored authority catalog assertion failed")
	}
	return signed, true, nil
}

func isNoRows(err error) bool { return err == pgx.ErrNoRows }
