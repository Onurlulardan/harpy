package main

import (
	"bytes"
	"context"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

func loadPreviousCheckpoint(
	ctx context.Context, query catalogQuerier,
	operation cutoverv1.AuthorityAdoptionOperation,
) (*cutoverv1.AuthorityCatalogCheckpoint, error) {
	var count int
	var generation *int
	if err := query.QueryRow(ctx, `SELECT count(*)::integer, max(catalog_generation)::integer
FROM cutover_authority.authority_adoption_inventory`).Scan(&count, &generation); err != nil {
		return nil, fmt.Errorf("authority adoption inventory read failed")
	}
	if operation.Mode == "initial_adoption" {
		if count != 0 || generation != nil {
			return nil, fmt.Errorf("initial authority adoption inventory is not empty")
		}
		return nil, nil
	}
	if generation == nil || *generation != operation.PreviousCatalogGeneration ||
		count != operation.PreviousCatalogGeneration {
		return nil, fmt.Errorf("prior authority adoption inventory differs")
	}
	var document []byte
	var digest string
	err := query.QueryRow(ctx, `SELECT receipt_document, receipt_sha256
FROM cutover_authority.authority_adoption_inventory WHERE catalog_generation=$1`,
		operation.PreviousCatalogGeneration).Scan(&document, &digest)
	if err != nil || operation.PreviousAuthorityReceiptSHA256 == nil ||
		digest != *operation.PreviousAuthorityReceiptSHA256 {
		return nil, fmt.Errorf("prior authority receipt differs")
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil || !bytes.Equal(document, canonical) {
		return nil, fmt.Errorf("prior authority receipt is not canonical")
	}
	computed, err := attestation.DigestHex(document)
	if err != nil || computed != digest {
		return nil, fmt.Errorf("prior authority receipt digest differs")
	}
	var receipt cutoverv1.AuthorityAdoptionReceipt
	if err := attestation.DecodeStrict(document, &receipt); err != nil || receipt.Validate() != nil {
		return nil, fmt.Errorf("prior authority receipt contract differs")
	}
	return &cutoverv1.AuthorityCatalogCheckpoint{
		ReceiptSHA256: digest, ReceiptID: receipt.ReceiptID,
		EnvironmentID: receipt.EnvironmentID, Target: receipt.Target,
		CatalogGeneration: receipt.CatalogGeneration, AdoptedAt: receipt.AdoptedAt,
		CatalogContractSHA256: receipt.PostCatalogContractSHA256,
		CatalogInstanceSHA256: receipt.CatalogInstanceSHA256,
		Objects:               append([]cutoverv1.AuthorityAdoptionReceiptObject(nil), receipt.Objects...),
	}, nil
}

func insertAdoptionReceipt(
	ctx context.Context, transaction pgx.Tx, signed signedAdoptionReceipt,
) error {
	receipt := signed.receipt
	_, err := transaction.Exec(ctx, `INSERT INTO cutover_authority.authority_adoption_inventory (
    catalog_generation,operation_id,receipt_id,receipt_sha256,receipt_document,
    operation_sha256,release_id,environment_id,database_origin,system_identifier,
    database_oid,catalog_contract_sha256,catalog_instance_sha256,
    dependency_contract_sha256,dependency_instance_sha256,object_count,
    dependency_count,adopted_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::numeric,$11::oid,$12,$13,$14,$15,$16,$17,$18)`,
		receipt.CatalogGeneration, receipt.OperationID, receipt.ReceiptID,
		signed.digest, signed.document, receipt.OperationSHA256, receipt.ReleaseID,
		receipt.EnvironmentID, receipt.Target.DatabaseOrigin,
		receipt.Target.SystemIdentifier, receipt.Target.DatabaseOID,
		receipt.PostCatalogContractSHA256, receipt.CatalogInstanceSHA256,
		receipt.DependencyContractSHA256, receipt.DependencyInstanceSHA256,
		len(receipt.Objects), len(receipt.Dependencies), receipt.AdoptedAt,
	)
	if err != nil {
		return fmt.Errorf("protected authority inventory insert failed")
	}
	return nil
}
