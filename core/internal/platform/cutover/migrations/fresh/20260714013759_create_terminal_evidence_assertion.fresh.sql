-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.assert_terminal_evidence_v1(
    p_operation_id uuid, p_run_id uuid, p_terminal_state text, p_terminal_evidence_document bytea
)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_terminal jsonb;
BEGIN
    IF SESSION_USER NOT IN ('forgecrud_cluster_admin','forgecrud_cutover_finalizer') OR
       p_operation_id IS NULL OR p_run_id IS NULL OR p_terminal_state NOT IN ('completed','failed','cancelled') OR
       p_terminal_evidence_document IS NULL OR pg_catalog.octet_length(p_terminal_evidence_document) NOT BETWEEN 2 AND 16777216 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover terminal evidence input is invalid';
    END IF;
    v_terminal:=pg_catalog.convert_from(p_terminal_evidence_document,'UTF8')::jsonb;
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    IF pg_catalog.jsonb_typeof(v_terminal) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_terminal)) IS DISTINCT FROM
         (CASE WHEN p_terminal_state='completed' THEN 14::bigint ELSE 9::bigint END) OR
       (v_terminal ?& ARRAY['contract_version','operation_id','run_id','lease_manifest_sha256','plan_sha256','transform_manifest_sha256','terminal_state']) IS DISTINCT FROM true OR
       pg_catalog.jsonb_typeof(v_terminal->'contract_version') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_terminal->'operation_id') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_terminal->'run_id') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_terminal->'lease_manifest_sha256') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_terminal->'plan_sha256') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_terminal->'transform_manifest_sha256') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_terminal->'terminal_state') IS DISTINCT FROM 'string' OR
       v_terminal->>'contract_version' IS DISTINCT FROM 'cutover-terminal-evidence.v1' OR
       v_terminal->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       v_terminal->>'run_id' IS DISTINCT FROM p_run_id::text OR
       v_terminal->>'lease_manifest_sha256' IS DISTINCT FROM v_operation.lease_manifest_sha256 OR
       v_terminal->>'plan_sha256' IS DISTINCT FROM v_operation.plan_sha256 OR
       v_terminal->>'transform_manifest_sha256' IS DISTINCT FROM v_operation.transform_manifest_sha256 OR
       v_terminal->>'terminal_state' IS DISTINCT FROM p_terminal_state OR
       (p_terminal_state='completed' AND v_operation.state IS DISTINCT FROM 'quiescing') OR
       (p_terminal_state IN ('failed','cancelled') AND v_operation.state NOT IN ('planned','copying','shadow','quiescing','blocked')) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover terminal evidence identity differs';
    END IF;
    IF p_terminal_state IN ('failed','cancelled') THEN
        IF (v_terminal ?& ARRAY['outcome_code','outcome_evidence_sha256']) IS DISTINCT FROM true OR
           pg_catalog.jsonb_typeof(v_terminal->'outcome_code') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_terminal->'outcome_evidence_sha256') IS DISTINCT FROM 'string' OR
           v_terminal->>'outcome_evidence_sha256'!~'^[0-9a-f]{64}$' OR
           (p_terminal_state='failed' AND NOT (v_terminal->>'outcome_code'=ANY
             (ARRAY['source_failed','target_failed','verification_failed','lease_expired']::text[]))) OR
           (p_terminal_state='cancelled' AND NOT (v_terminal->>'outcome_code'=ANY
             (ARRAY['operator_cancelled','superseded']::text[]))) THEN
            RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover abort evidence differs';
        END IF;
        RETURN;
    END IF;
    IF (v_terminal ?& ARRAY['final_lsn','initial_copy','checkpoint_count','checkpoints','journal_transaction_count','applied_change_count','journal_coverage_root_sha256']) IS DISTINCT FROM true OR
       pg_catalog.jsonb_typeof(v_terminal->'final_lsn') IS DISTINCT FROM 'string' OR
       (v_terminal->>'final_lsn')::pg_lsn::text IS DISTINCT FROM v_terminal->>'final_lsn' OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_terminal->'initial_copy')) IS DISTINCT FROM 7::bigint OR
       (v_terminal->'initial_copy' ?& ARRAY['manifest_sha256','snapshot_name','snapshot_lsn','relation_count','partition_count','row_count','root_sha256']) IS DISTINCT FROM true OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'manifest_sha256') IS DISTINCT FROM 'string' OR
       v_terminal->'initial_copy'->>'manifest_sha256' IS DISTINCT FROM v_operation.lease_manifest_sha256 OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'snapshot_name') IS DISTINCT FROM 'string' OR
       v_terminal->'initial_copy'->>'snapshot_name'!~'^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$' OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'snapshot_lsn') IS DISTINCT FROM 'string' OR
       (v_terminal->'initial_copy'->>'snapshot_lsn')::pg_lsn::text IS DISTINCT FROM v_terminal->'initial_copy'->>'snapshot_lsn' OR
       (v_terminal->'initial_copy'->>'snapshot_lsn')::pg_lsn>(v_terminal->>'final_lsn')::pg_lsn OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'relation_count') IS DISTINCT FROM 'number' OR
       v_terminal->'initial_copy'->>'relation_count'!~'^[1-9][0-9]*$' OR
       (v_terminal->'initial_copy'->>'relation_count')::integer NOT BETWEEN 1 AND 4096 OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'partition_count') IS DISTINCT FROM 'number' OR
       v_terminal->'initial_copy'->>'partition_count'!~'^[1-9][0-9]*$' OR
       (v_terminal->'initial_copy'->>'partition_count')::integer NOT BETWEEN
         (v_terminal->'initial_copy'->>'relation_count')::integer AND 10000 OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'row_count') IS DISTINCT FROM 'number' OR
       v_terminal->'initial_copy'->>'row_count'!~'^[0-9]+$' OR
       (v_terminal->'initial_copy'->>'row_count')::numeric>9007199254740991 OR
       pg_catalog.jsonb_typeof(v_terminal->'initial_copy'->'root_sha256') IS DISTINCT FROM 'string' OR
       v_terminal->'initial_copy'->>'root_sha256'!~'^[0-9a-f]{64}$' OR
       pg_catalog.jsonb_typeof(v_terminal->'checkpoint_count') IS DISTINCT FROM 'number' OR
       v_terminal->>'checkpoint_count'!~'^[1-9][0-9]*$' OR
       (v_terminal->>'checkpoint_count')::integer NOT BETWEEN 1 AND 10000 OR
       pg_catalog.jsonb_typeof(v_terminal->'checkpoints') IS DISTINCT FROM 'array' OR
       (v_terminal->>'checkpoint_count')::integer IS DISTINCT FROM pg_catalog.jsonb_array_length(v_terminal->'checkpoints') OR
       (v_terminal->>'checkpoint_count')::integer IS DISTINCT FROM (v_terminal->'initial_copy'->>'partition_count')::integer OR
       pg_catalog.jsonb_typeof(v_terminal->'journal_transaction_count') IS DISTINCT FROM 'number' OR
       v_terminal->>'journal_transaction_count'!~'^[0-9]+$' OR
       (v_terminal->>'journal_transaction_count')::numeric>9007199254740991 OR
       pg_catalog.jsonb_typeof(v_terminal->'applied_change_count') IS DISTINCT FROM 'number' OR
       v_terminal->>'applied_change_count'!~'^[0-9]+$' OR
       (v_terminal->>'applied_change_count')::numeric>9007199254740991 OR
       ((v_terminal->>'journal_transaction_count')::numeric=0) IS DISTINCT FROM
         ((v_terminal->>'applied_change_count')::numeric=0) OR
       pg_catalog.jsonb_typeof(v_terminal->'journal_coverage_root_sha256') IS DISTINCT FROM 'string' OR
       v_terminal->>'journal_coverage_root_sha256'!~'^[0-9a-f]{64}$' OR
       EXISTS (SELECT 1 FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') AS item
          WHERE pg_catalog.jsonb_typeof(item) IS DISTINCT FROM 'object'
             OR (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(item)) IS DISTINCT FROM 6::bigint
             OR (item ?& ARRAY['relation_ordinal','checkpoint_key','applied_lsn','applied_rows','batch_count','batch_root_sha256']) IS DISTINCT FROM true
             OR pg_catalog.jsonb_typeof(item->'relation_ordinal') IS DISTINCT FROM 'number'
             OR item->>'relation_ordinal'!~'^[1-9][0-9]*$' OR (item->>'relation_ordinal')::integer NOT BETWEEN 1 AND 4096
             OR pg_catalog.jsonb_typeof(item->'checkpoint_key') IS DISTINCT FROM 'string'
             OR pg_catalog.octet_length(item->>'checkpoint_key') NOT BETWEEN 1 AND 512 OR item->>'checkpoint_key'~'[[:cntrl:]]'
             OR pg_catalog.jsonb_typeof(item->'applied_lsn') IS DISTINCT FROM 'string'
             OR (item->>'applied_lsn')::pg_lsn::text IS DISTINCT FROM item->>'applied_lsn'
             OR (item->>'applied_lsn')::pg_lsn<(v_terminal->>'final_lsn')::pg_lsn
             OR pg_catalog.jsonb_typeof(item->'applied_rows') IS DISTINCT FROM 'number'
             OR item->>'applied_rows'!~'^[0-9]+$' OR (item->>'applied_rows')::numeric>9007199254740991
             OR pg_catalog.jsonb_typeof(item->'batch_count') IS DISTINCT FROM 'number'
             OR item->>'batch_count'!~'^[1-9][0-9]*$' OR (item->>'batch_count')::numeric>2147483647
             OR pg_catalog.jsonb_typeof(item->'batch_root_sha256') IS DISTINCT FROM 'string'
             OR item->>'batch_root_sha256'!~'^[0-9a-f]{64}$') OR
       EXISTS (SELECT 1 FROM (SELECT item->>'relation_ordinal' AS relation_ordinal,item->>'checkpoint_key' AS checkpoint_key,
          pg_catalog.lag(item->>'relation_ordinal') OVER (ORDER BY item_ordinal) AS prior_relation_ordinal,
          pg_catalog.lag(item->>'checkpoint_key') OVER (ORDER BY item_ordinal) AS prior_checkpoint_key,item_ordinal
          FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') WITH ORDINALITY AS value(item,item_ordinal)) AS ordered
          WHERE ordered.item_ordinal>1 AND ((ordered.relation_ordinal)::integer<(ordered.prior_relation_ordinal)::integer
             OR ((ordered.relation_ordinal)::integer=(ordered.prior_relation_ordinal)::integer
                AND ordered.checkpoint_key<=ordered.prior_checkpoint_key))) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover completion evidence shape differs';
    END IF;
    PERFORM cutover_authority.assert_completed_evidence_v1(p_operation_id,p_run_id,p_terminal_evidence_document);
    PERFORM cutover_authority.assert_tenant_completion_v1(p_operation_id,p_run_id);
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.assert_terminal_evidence_v1(uuid,uuid,text,bytea) FROM PUBLIC;
