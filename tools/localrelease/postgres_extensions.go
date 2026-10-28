package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func reconcileExtensions(ctx context.Context, target *pgx.Conn) error {
	if _, err := target.Exec(ctx,
		"CREATE SCHEMA IF NOT EXISTS platform_ext AUTHORIZATION harpy_owner",
	); err != nil {
		return fmt.Errorf("create extension schema: %w", err)
	}
	var schemaOwner string
	if err := target.QueryRow(ctx, `SELECT owner.rolname FROM pg_catalog.pg_namespace AS schema
JOIN pg_catalog.pg_roles AS owner ON owner.oid=schema.nspowner
WHERE schema.nspname='platform_ext'`).Scan(&schemaOwner); err != nil {
		return fmt.Errorf("inspect extension schema: %w", err)
	}
	if schemaOwner != "harpy_owner" {
		return fmt.Errorf("platform_ext schema owner differs")
	}
	tx, err := target.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_extension_admin"); err != nil {
		return fmt.Errorf("assume extension admin: %w", err)
	}
	statements := []string{
		"CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA platform_ext VERSION '1.6'",
		"CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA platform_ext VERSION '0.8.3'",
	}
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("reconcile extension: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit extension reconciliation: %w", err)
	}
	if _, err := target.Exec(ctx, "REVOKE CREATE ON SCHEMA public FROM PUBLIC"); err != nil {
		return fmt.Errorf("revoke public schema create: %w", err)
	}
	return verifyExtensions(ctx, target)
}

func verifyExtensions(ctx context.Context, target *pgx.Conn) error {
	for _, expected := range extensionManifest().Extensions {
		var version, schema, owner string
		err := target.QueryRow(ctx, `SELECT extension.extversion, schema.nspname, owner.rolname
FROM pg_catalog.pg_extension AS extension
JOIN pg_catalog.pg_namespace AS schema ON schema.oid=extension.extnamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=extension.extowner
WHERE extension.extname=$1`, expected.Name).Scan(&version, &schema, &owner)
		if err != nil {
			return fmt.Errorf("inspect extension %s: %w", expected.Name, err)
		}
		if version != expected.Version || schema != expected.Schema || owner != expected.Owner {
			return fmt.Errorf("extension %s catalog identity differs", expected.Name)
		}
		actualHash, err := readControlFileHash(ctx, target, expected.ControlFile)
		if err != nil {
			return err
		}
		if actualHash != expected.ControlFileSHA256 {
			return fmt.Errorf("extension %s control file digest differs", expected.Name)
		}
	}
	return nil
}

func readControlFileHash(ctx context.Context, target *pgx.Conn, file string) (string, error) {
	var contents []byte
	err := target.QueryRow(ctx, `SELECT pg_catalog.pg_read_binary_file(
(SELECT setting FROM pg_catalog.pg_config WHERE name='SHAREDIR') || '/extension/' || $1
)`, file).Scan(&contents)
	if err != nil {
		return "", fmt.Errorf("read extension control file %s: %w", file, err)
	}
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:]), nil
}
