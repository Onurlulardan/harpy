-- +goose Up
REVOKE ALL ON FUNCTION catalog_guard_purge_approval_v2() FROM PUBLIC;

