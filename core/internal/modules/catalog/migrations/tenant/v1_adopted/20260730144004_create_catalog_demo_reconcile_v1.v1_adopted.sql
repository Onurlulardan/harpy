-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_reconcile_v1(
    p_receipt_id uuid, p_occurred_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_provision catalog_demo_provisions%ROWTYPE;
    v_handoff catalog_demo_handoffs%ROWTYPE;
BEGIN
    SELECT * INTO STRICT v_provision
    FROM catalog_demo_provisions AS provision
    WHERE provision.receipt_id = p_receipt_id
    FOR UPDATE;
    SELECT * INTO v_handoff
    FROM catalog_demo_handoffs AS handoff
    WHERE handoff.provision_receipt_id = p_receipt_id
    FOR SHARE;
    IF FOUND THEN
        IF v_handoff.completed_at < v_provision.created_at THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Catalog demo handoff predates provision';
        END IF;
        IF NOT v_provision.ready THEN
            UPDATE catalog_demo_provisions
            SET ready = true,
                retry_at = NULL,
                completed_at = v_handoff.completed_at,
                updated_at = GREATEST(p_occurred_at, v_handoff.completed_at)
            WHERE receipt_id = p_receipt_id;
        END IF;
        RETURN;
    END IF;
    IF v_provision.ready THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo ready receipt lacks handoff';
    END IF;
    IF p_occurred_at >= v_provision.retry_at THEN
        UPDATE catalog_demo_provisions
        SET retry_at = p_occurred_at + interval '30 seconds',
            completed_at = p_occurred_at,
            updated_at = p_occurred_at
        WHERE receipt_id = p_receipt_id;
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_demo_reconcile_v1(uuid,timestamptz) FROM PUBLIC;
