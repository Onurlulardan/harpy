-- +goose Up
CREATE TABLE public.tenants (
    id uuid PRIMARY KEY DEFAULT pg_catalog.gen_random_uuid(),
    name varchar(255) NOT NULL,
    slug varchar(100) NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    settings jsonb DEFAULT '{}'::jsonb,
    owner_id uuid NOT NULL,
    schema_name text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);
