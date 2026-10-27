-- +goose Up
-- Migration stream version: 20260729100012.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_successor_finalize_transition_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY INVOKER SET search_path=pg_catalog
AS $function$
DECLARE
    v_receipt cutover_authority.finalization_receipts%ROWTYPE;
    v_companion platform.cutover_operations%ROWTYPE;
    v_companion_receipt cutover_authority.finalization_receipts%ROWTYPE;
    v_verified boolean;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'
       OR TG_TABLE_NAME IS DISTINCT FROM 'cutover_operations' THEN
        RAISE EXCEPTION USING ERRCODE='P5740',
            MESSAGE='cutover successor finalizer trigger target differs';
    END IF;
    IF OLD.state IS DISTINCT FROM 'quiescing' OR NEW.state IS DISTINCT FROM 'completed'
       OR NEW.group_id NOT IN ('analytics','billing','billing_forms','workflow') THEN
        RETURN NEW;
    END IF;
    IF NEW.scope_kind NOT IN ('global','tenant')
       OR (NEW.scope_kind='global'
           AND (NEW.group_id<>'billing' OR NEW.domain<>'billing'))
       OR (NEW.scope_kind='tenant' AND NEW.group_id='billing')
       OR NEW.operation_id IS DISTINCT FROM OLD.operation_id
       OR NEW.run_id IS DISTINCT FROM OLD.run_id
       OR NEW.scope_id IS DISTINCT FROM OLD.scope_id
       OR NEW.group_id IS DISTINCT FROM OLD.group_id
       OR NEW.domain IS DISTINCT FROM OLD.domain
       OR NEW.cutover_generation IS DISTINCT FROM OLD.cutover_generation THEN
        RAISE EXCEPTION USING ERRCODE='P5741',
            MESSAGE='cutover successor completed transition differs';
    END IF;
    SELECT receipt.* INTO STRICT v_receipt
    FROM cutover_authority.finalization_receipts AS receipt
    WHERE receipt.operation_id=NEW.operation_id AND receipt.run_id=NEW.run_id
      AND receipt.terminal_state='completed';
    SELECT cutover_authority.verify_finalization_receipt_v1(
        NEW.operation_id,NEW.run_id,v_receipt.finalizer_lease_id,
        'completed',v_receipt.terminal_evidence_sha256) INTO STRICT v_verified;
    IF NOT v_verified THEN
        RAISE EXCEPTION USING ERRCODE='P5742',
            MESSAGE='cutover successor protected receipt differs';
    END IF;
    IF NEW.group_id='billing_forms' THEN
        PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
            NEW.scope_kind||':'||NEW.scope_id::text||':'||
            NEW.group_id||':'||NEW.cutover_generation::text,0));
        PERFORM 1 FROM platform.cutover_groups AS group_row
        WHERE group_row.scope_kind=NEW.scope_kind
          AND group_row.scope_id=NEW.scope_id
          AND group_row.group_id=NEW.group_id
          AND group_row.generation=NEW.cutover_generation
          AND group_row.state='quiescing' AND group_row.writer='none';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P5744',
                MESSAGE='billing forms serialization authority differs';
        END IF;
        SELECT operation.* INTO v_companion
        FROM platform.cutover_operations AS operation
        WHERE operation.scope_kind=NEW.scope_kind AND operation.scope_id=NEW.scope_id
          AND operation.group_id=NEW.group_id AND operation.state='completed'
          AND operation.cutover_generation=NEW.cutover_generation
          AND operation.domain=CASE NEW.domain
              WHEN 'billing' THEN 'forms' ELSE 'billing' END
        ORDER BY operation.operation_id,operation.run_id LIMIT 1 FOR SHARE;
        IF NOT FOUND THEN RETURN NEW; END IF;
        IF EXISTS (
            SELECT 1 FROM platform.cutover_operations AS duplicate
            WHERE duplicate.scope_kind=NEW.scope_kind AND duplicate.scope_id=NEW.scope_id
              AND duplicate.group_id=NEW.group_id AND duplicate.state='completed'
              AND duplicate.cutover_generation=NEW.cutover_generation
              AND duplicate.domain=v_companion.domain
              AND ROW(duplicate.operation_id,duplicate.run_id) IS DISTINCT FROM
                  ROW(v_companion.operation_id,v_companion.run_id)
        ) THEN
            RAISE EXCEPTION USING ERRCODE='P5743',
                MESSAGE='billing forms companion operation is ambiguous';
        END IF;
        SELECT receipt.* INTO STRICT v_companion_receipt
        FROM cutover_authority.finalization_receipts AS receipt
        WHERE receipt.operation_id=v_companion.operation_id
          AND receipt.run_id=v_companion.run_id AND receipt.terminal_state='completed';
        SELECT cutover_authority.verify_finalization_receipt_v1(
            v_companion.operation_id,v_companion.run_id,
            v_companion_receipt.finalizer_lease_id,'completed',
            v_companion_receipt.terminal_evidence_sha256) INTO STRICT v_verified;
        IF NOT v_verified OR v_companion_receipt.terminal_evidence_sha256
           IS DISTINCT FROM v_receipt.terminal_evidence_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P5742',
                MESSAGE='billing forms protected receipts differ';
        END IF;
        PERFORM platform.cutover_finalize_successor_group_v1(
            NEW.operation_id,NEW.run_id,v_companion.operation_id,
            v_companion.run_id,v_receipt.terminal_evidence_sha256);
    ELSE
        PERFORM platform.cutover_finalize_successor_group_v1(
            NEW.operation_id,NEW.run_id,NULL,NULL,
            v_receipt.terminal_evidence_sha256);
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_successor_finalize_transition_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.cutover_successor_finalize_transition_v1()
    TO forgecrud_cutover_authority_owner;
CREATE TRIGGER cutover_operations_successor_finalized_switch
AFTER UPDATE OF state ON platform.cutover_operations FOR EACH ROW
EXECUTE FUNCTION platform.cutover_successor_finalize_transition_v1();
ALTER TABLE platform.cutover_operations ENABLE ALWAYS TRIGGER
    cutover_operations_successor_finalized_switch;
