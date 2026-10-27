-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.notifications_retention_purge_platform_v1(
    p_tenant_id uuid, p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $$
DECLARE v_count bigint;
BEGIN
WITH terminal_tickets AS MATERIALIZED (
    SELECT ticket.id, ticket.tenant_id, ticket.user_id,
           ticket.authorization_version, ticket.origin,
           CASE WHEN ticket.consumed_at IS NULL THEN 'expired' ELSE 'consumed' END AS kind,
           COALESCE(ticket.consumed_at, ticket.expires_at) AS terminal_at
    FROM platform.notification_realtime_tickets AS ticket
    WHERE p_tenant_id IS NOT NULL AND pg_catalog.isfinite(p_now)
      AND p_limit BETWEEN 1 AND 500 AND ticket.tenant_id = p_tenant_id
      AND COALESCE(ticket.consumed_at, ticket.expires_at) <= p_now - interval '24 hours'
    ORDER BY COALESCE(ticket.consumed_at, ticket.expires_at), ticket.id
    FOR UPDATE SKIP LOCKED LIMIT p_limit
), inserted_audits AS (
    INSERT INTO platform.notification_realtime_ticket_audits (
        ticket_id, tenant_id, user_id, authorization_version, origin_digest,
        terminal_kind, terminal_at, retain_until
    ) SELECT id, tenant_id, user_id, authorization_version,
        pg_catalog.sha256(pg_catalog.convert_to(origin, 'UTF8')),
        kind, terminal_at, terminal_at + interval '90 days'
    FROM terminal_tickets ON CONFLICT (ticket_id) DO NOTHING RETURNING ticket_id
), removable_tickets AS (
    SELECT candidate.id FROM terminal_tickets AS candidate
    WHERE candidate.id IN (SELECT ticket_id FROM inserted_audits)
       OR EXISTS (
          SELECT 1 FROM platform.notification_realtime_ticket_audits AS audit
          WHERE audit.ticket_id = candidate.id
            AND audit.tenant_id = candidate.tenant_id
            AND audit.user_id = candidate.user_id
            AND audit.authorization_version = candidate.authorization_version
            AND audit.origin_digest = pg_catalog.sha256(
                pg_catalog.convert_to(candidate.origin, 'UTF8'))
            AND audit.terminal_kind = candidate.kind
            AND audit.terminal_at = candidate.terminal_at
       )
), deleted_tickets AS (
    DELETE FROM platform.notification_realtime_tickets AS ticket
    USING removable_tickets WHERE ticket.id = removable_tickets.id RETURNING ticket.id
), locator_candidates AS MATERIALIZED (
    SELECT provider_kind, provider_account_digest, external_message_digest
    FROM platform.notification_provider_locators
    WHERE tenant_id = p_tenant_id AND expires_at <= p_now
    ORDER BY expires_at, delivery_id FOR UPDATE SKIP LOCKED LIMIT p_limit
), deleted_locators AS (
    DELETE FROM platform.notification_provider_locators AS locator
    USING locator_candidates AS candidate
    WHERE locator.provider_kind = candidate.provider_kind
      AND locator.provider_account_digest = candidate.provider_account_digest
      AND locator.external_message_digest = candidate.external_message_digest
    RETURNING locator.delivery_id
), audit_candidates AS MATERIALIZED (
    SELECT ticket_id FROM platform.notification_realtime_ticket_audits
    WHERE tenant_id = p_tenant_id AND retain_until <= p_now
    ORDER BY retain_until, ticket_id FOR UPDATE SKIP LOCKED LIMIT p_limit
), deleted_audits AS (
    DELETE FROM platform.notification_realtime_ticket_audits AS audit
    USING audit_candidates AS candidate
    WHERE audit.ticket_id = candidate.ticket_id RETURNING audit.ticket_id
), legacy_candidates AS MATERIALIZED (
    SELECT id FROM platform.notification_legacy_connection_hits
    WHERE tenant_id = p_tenant_id AND occurred_at <= p_now - interval '90 days'
    ORDER BY occurred_at, id FOR UPDATE SKIP LOCKED LIMIT p_limit
), deleted_legacy AS (
    DELETE FROM platform.notification_legacy_connection_hits AS hit
    USING legacy_candidates AS candidate
    WHERE hit.id = candidate.id RETURNING hit.id
)
SELECT (SELECT pg_catalog.count(*) FROM deleted_tickets)
     + (SELECT pg_catalog.count(*) FROM deleted_locators)
     + (SELECT pg_catalog.count(*) FROM deleted_audits)
     + (SELECT pg_catalog.count(*) FROM deleted_legacy) INTO v_count;
RETURN v_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.notifications_retention_purge_platform_v1(
    uuid, timestamptz, integer
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.notifications_retention_purge_platform_v1(
    uuid, timestamptz, integer
) TO harpy_runtime;
