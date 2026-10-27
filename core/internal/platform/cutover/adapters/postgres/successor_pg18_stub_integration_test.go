//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

const successorAcceptanceDigest = "44136fa355b3678a1146ad16f7e8649e" +
	"94fb4fc21fe77e8310c060f61caaff8a"

func prepareSuccessorModuleStubs(
	t *testing.T,
	ctx context.Context,
	connection *pgx.Conn,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.successor_authority_test_calls (
    module text NOT NULL,
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    companion_operation_id uuid,
    companion_run_id uuid,
    acceptance_digest text NOT NULL,
    authority_id uuid NOT NULL,
    call_xid xid8 NOT NULL DEFAULT pg_catalog.pg_current_xact_id(),
    PRIMARY KEY (module,operation_id,run_id)
);
CREATE FUNCTION platform.analytics_finalize_authority_successor_v1(
    p_operation_id uuid,p_run_id uuid,p_digest text
) RETURNS uuid LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_id uuid;
BEGIN
    IF CURRENT_USER<>'harpy_owner' OR SESSION_USER<>'harpy_cutover_finalizer' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='analytics successor stub denied';
    END IF;
    SELECT authority_id INTO v_id FROM platform.successor_authority_test_calls
    WHERE module='analytics' AND operation_id=p_operation_id AND run_id=p_run_id;
    IF FOUND THEN RETURN v_id; END IF;
    v_id:=pg_catalog.uuidv7();
    INSERT INTO platform.successor_authority_test_calls
        (module,operation_id,run_id,acceptance_digest,authority_id)
    VALUES ('analytics',p_operation_id,p_run_id,p_digest,v_id);
    RETURN v_id;
END;$function$;
CREATE FUNCTION platform.workflow_finalize_authority_successor_v1(
    p_operation_id uuid,p_run_id uuid,p_digest text
) RETURNS uuid LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_id uuid;
BEGIN
    IF CURRENT_USER<>'harpy_owner' OR SESSION_USER<>'harpy_cutover_finalizer' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='workflow successor stub denied';
    END IF;
    SELECT authority_id INTO v_id FROM platform.successor_authority_test_calls
    WHERE module='workflow' AND operation_id=p_operation_id AND run_id=p_run_id;
    IF FOUND THEN RETURN v_id; END IF;
    v_id:=pg_catalog.uuidv7();
    INSERT INTO platform.successor_authority_test_calls
        (module,operation_id,run_id,acceptance_digest,authority_id)
    VALUES ('workflow',p_operation_id,p_run_id,p_digest,v_id);
    RETURN v_id;
END;$function$;
CREATE FUNCTION platform.billing_finalize_authority_successor_v1(
    p_operation_id uuid,p_run_id uuid,p_digest text
) RETURNS uuid LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_id uuid;
BEGIN
    IF CURRENT_USER<>'harpy_owner' OR SESSION_USER<>'harpy_cutover_finalizer' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='billing successor stub denied';
    END IF;
    SELECT authority_id INTO v_id FROM platform.successor_authority_test_calls
    WHERE module='billing' AND operation_id=p_operation_id AND run_id=p_run_id;
    IF FOUND THEN RETURN v_id; END IF;
    v_id:=pg_catalog.uuidv7();
    INSERT INTO platform.successor_authority_test_calls (
        module,operation_id,run_id,acceptance_digest,authority_id
    ) VALUES ('billing',p_operation_id,p_run_id,p_digest,v_id);
    RETURN v_id;
END;$function$;
REVOKE ALL ON FUNCTION platform.analytics_finalize_authority_successor_v1(
    uuid,uuid,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.workflow_finalize_authority_successor_v1(
    uuid,uuid,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.billing_finalize_authority_successor_v1(
    uuid,uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_finalize_authority_successor_v1(
    uuid,uuid,text) TO harpy_cutover_authority_owner;
GRANT EXECUTE ON FUNCTION platform.workflow_finalize_authority_successor_v1(
    uuid,uuid,text) TO harpy_cutover_authority_owner;
GRANT EXECUTE ON FUNCTION platform.billing_finalize_authority_successor_v1(
    uuid,uuid,text) TO harpy_cutover_authority_owner;`
	if _, err := connection.Exec(ctx, statement); err != nil {
		t.Fatalf("prepare successor module stubs: %v", err)
	}
}

func attachSuccessorReceiptHarness(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
) {
	t.Helper()
	const statement = `
CREATE TABLE cutover_authority.successor_test_receipt_permissions (
    operation_id uuid NOT NULL,run_id uuid NOT NULL,lease_id uuid NOT NULL,
    digest text NOT NULL,PRIMARY KEY(operation_id,run_id)
);
ALTER TABLE cutover_authority.successor_test_receipt_permissions
    OWNER TO harpy_owner;
CREATE OR REPLACE FUNCTION cutover_authority.verify_finalization_receipt_v1(
    p_operation_id uuid,p_run_id uuid,p_finalizer_lease_id uuid,
    p_terminal_state text,p_terminal_evidence_sha256 text
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF SESSION_USER<>'harpy_cutover_finalizer' OR
       p_terminal_state<>'completed' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='successor verifier denied';
    END IF;
    RETURN EXISTS (
        SELECT 1 FROM cutover_authority.successor_test_receipt_permissions AS allowed
        WHERE allowed.operation_id=p_operation_id AND allowed.run_id=p_run_id
          AND allowed.lease_id=p_finalizer_lease_id
          AND allowed.digest=p_terminal_evidence_sha256
    );
END;$function$;
CREATE FUNCTION cutover_authority.successor_test_finalize_v1(
    p_operation_id uuid,p_run_id uuid,p_lease_id uuid,p_digest text
) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_doc bytea:=pg_catalog.convert_to('{}','UTF8');
    v_catalog bytea:=pg_catalog.convert_to(
      '{"contract_version":"tenant-migration-requirement-catalog.v2","requirements":[{}],"targets":[]}',
      'UTF8');
    v_sha text:=pg_catalog.encode(pg_catalog.sha256(v_doc),'hex');
BEGIN
    IF SESSION_USER<>'harpy_cutover_finalizer' OR
       CURRENT_USER<>'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='successor test finalizer denied';
    END IF;
    INSERT INTO cutover_authority.finalization_receipts (
      operation_id,run_id,finalizer_lease_id,finalizer_lease_sha256,
      finalizer_lease_document,adoption_lease_sha256,adoption_lease_document,
      seal_receipt_sha256,seal_receipt_document,cleanup_receipt_sha256,
      cleanup_receipt_document,authority_receipt_sha256,authority_receipt_document,
      terminal_state,terminal_evidence_sha256,terminal_evidence_document,
      finalization_state_sha256,tenant_requirement_catalog_sha256,
      tenant_requirement_catalog_document,tenant_requirement_count,tenant_target_count,
      finalized_by
    ) VALUES (
      p_operation_id,p_run_id,p_lease_id,v_sha,v_doc,v_sha,v_doc,v_sha,v_doc,
      v_sha,v_doc,v_sha,v_doc,'completed',p_digest,v_doc,
      pg_catalog.repeat('e',64),pg_catalog.encode(pg_catalog.sha256(v_catalog),'hex'),
      v_catalog,1,0,'harpy_cutover_finalizer'
    ) ON CONFLICT DO NOTHING;
    UPDATE platform.cutover_operations SET state='completed',
      updated_at=pg_catalog.clock_timestamp()
    WHERE operation_id=p_operation_id AND run_id=p_run_id AND state='quiescing';
END;$function$;
ALTER FUNCTION cutover_authority.successor_test_finalize_v1(uuid,uuid,uuid,text)
    OWNER TO harpy_owner;
REVOKE ALL ON FUNCTION
    cutover_authority.successor_test_finalize_v1(uuid,uuid,uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    cutover_authority.successor_test_finalize_v1(uuid,uuid,uuid,text)
    TO harpy_cutover_finalizer;`
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("attach successor receipt harness: %v", err)
	}
}
