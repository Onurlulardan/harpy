-- +goose Up
ALTER TABLE public.tenants
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN status DROP DEFAULT;
