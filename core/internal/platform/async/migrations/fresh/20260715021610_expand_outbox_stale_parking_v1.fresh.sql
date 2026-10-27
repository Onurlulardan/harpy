-- +goose Up
ALTER TABLE platform.outbox
    ADD COLUMN parked_at timestamptz,
    ADD COLUMN parked_code text
        CHECK (parked_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    ADD COLUMN parked_current_group_generation bigint
        CHECK (parked_current_group_generation > 0),
    ADD COLUMN parked_current_cutover_generation bigint
        CHECK (parked_current_cutover_generation > 0),
    ADD CONSTRAINT outbox_parked_check CHECK (
        (parked_at IS NULL AND parked_code IS NULL AND
            parked_current_group_generation IS NULL AND
            parked_current_cutover_generation IS NULL) OR
        (delivery_state = 'pending' AND parked_at IS NOT NULL AND
            parked_code = 'OUTBOX_AUTHORITY_STALE' AND
            lease_owner_id IS NULL AND lease_heartbeat_at IS NULL AND
            lease_until IS NULL AND
            parked_current_group_generation >= authority_group_generation AND
            parked_current_cutover_generation >= cutover_generation AND
            (parked_current_group_generation > authority_group_generation OR
             parked_current_cutover_generation > cutover_generation))
    );

CREATE INDEX outbox_stale_authority_idx ON platform.outbox (
    scope_kind, scope_id, authority_group_id, authority_domain,
    authority_group_generation, cutover_generation, created_at, event_id
) WHERE parked_at IS NULL AND delivery_state IN ('pending', 'publishing');
