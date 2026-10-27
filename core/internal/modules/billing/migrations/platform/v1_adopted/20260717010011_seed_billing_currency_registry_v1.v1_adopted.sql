-- +goose Up
INSERT INTO platform.billing_currency_registry_versions (
    authority_id, registry_version, source_published_on, source_url,
    source_bytes, source_sha256, normalized_sha256,
    unique_code_count, not_applicable_code_count, published_at
)
SELECT
    authority_id, 1, DATE '2026-01-01',
    'https://www.six-group.com/dam/download/financial-information/' ||
        'data-center/iso-currrency/lists/list-one.xml',
    47463,
    '838dfb991648cf36df939edd5fe3811737962b75a32252847d239cedd1e291c9',
    '894f71b85f203d76ef7dacb379caf4340638d7ce7c2a2ffafd0b5f65c1b6714f',
    178, 13, TIMESTAMPTZ '2026-01-01 00:00:00+00'
FROM platform.billing_global_authorities;

