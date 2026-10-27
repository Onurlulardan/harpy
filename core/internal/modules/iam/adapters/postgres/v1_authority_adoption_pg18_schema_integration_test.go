//go:build integration

package postgres

const v1AuthorityAdoptionSchema = `
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY, schema_name name NOT NULL UNIQUE,
    owner_id uuid NOT NULL, deleted_at timestamptz
);
CREATE TABLE platform.tenancy_deployment_policy (
    environment_id uuid PRIMARY KEY
);
CREATE TABLE platform.tenancy_tenant_bindings (
    environment_id uuid NOT NULL, tenant_id uuid NOT NULL, purpose text NOT NULL,
    PRIMARY KEY (environment_id,tenant_id)
);
CREATE TABLE platform.platform_cutover_v1_adoption_plans (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    decision text NOT NULL,
    primary_tenant_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    plan_sha256 text NOT NULL,
    plan_document jsonb NOT NULL
);
CREATE TABLE platform.test_cutover_batch_receipts (
    operation_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL,
    batch_ordinal integer NOT NULL,
    applied_count bigint NOT NULL,
    key_state_sha256 text NOT NULL,
    state_sha256 text NOT NULL,
    PRIMARY KEY (operation_id,relation_ordinal,batch_ordinal)
);
CREATE FUNCTION platform.cutover_prepare_batch_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,p_batch_ordinal integer,
    p_source_commit_lsn pg_lsn,p_source_xid xid8,p_batch_sha256 text,
    p_lease_manifest_sha256 text
) RETURNS TABLE(
    procedure_contract_id text,procedure_oid bigint,procedure_signature text,
    already_applied boolean,applied_count bigint,key_state_sha256 text,
    state_sha256 text
) LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    SELECT receipt.applied_count,receipt.key_state_sha256,receipt.state_sha256
      INTO applied_count,key_state_sha256,state_sha256
      FROM platform.test_cutover_batch_receipts AS receipt
     WHERE receipt.operation_id=p_operation_id
       AND receipt.relation_ordinal=p_relation_ordinal
       AND receipt.batch_ordinal=p_batch_ordinal;
    procedure_contract_id := 'iam.v1_protected_authority_adoption.v1';
    procedure_oid := 'platform.iam_v1_protected_authority_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid;
    procedure_signature := 'platform.iam_v1_protected_authority_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)';
    already_applied := FOUND;
    RETURN NEXT;
END;
$function$;
CREATE FUNCTION platform.cutover_record_batch_v1(
    p_operation_id uuid,p_run_id uuid,p_relation_ordinal smallint,
    p_checkpoint_key text,p_snapshot_name text,p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn,p_commit_lsn pg_lsn,p_final_lsn pg_lsn,
    p_source_xid xid8,p_batch_ordinal integer,p_checkpoint_state text,
    p_last_key jsonb,p_batch_sha256 text,p_batch_bytes bigint,
    p_applied_count bigint,p_key_state_sha256 text,p_state_sha256 text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    INSERT INTO platform.test_cutover_batch_receipts VALUES (
        p_operation_id,p_relation_ordinal,p_batch_ordinal,p_applied_count,
        p_key_state_sha256,p_state_sha256);
    RETURN true;
END;
$function$;
INSERT INTO platform.tenants VALUES (
    '71000000-0000-4000-8000-000000000002','t_v1_owner',
    '71000000-0000-4000-8000-000000000003',NULL
);
INSERT INTO platform.tenancy_deployment_policy VALUES (
    '71000000-0000-4000-8000-000000000001'
);
INSERT INTO platform.tenancy_tenant_bindings VALUES (
    '71000000-0000-4000-8000-000000000001',
    '71000000-0000-4000-8000-000000000002','onpremise_primary'
);`

const v1AuthorityDirectoryEvidence = `
INSERT INTO platform.iam_login_directory (
    tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
    source_identifier_id,source_revision,state,source_sha256,projected_at
) VALUES (
    '71000000-0000-4000-8000-000000000002','email_v1','v1-adoption',
    decode(repeat('1',64),'hex'),'71000000-0000-4000-8000-000000000003',
    '71000000-0000-4000-8000-000000000011',7,'active',
    decode(repeat('2',64),'hex'),'2026-07-23T00:00:00Z'
);`
