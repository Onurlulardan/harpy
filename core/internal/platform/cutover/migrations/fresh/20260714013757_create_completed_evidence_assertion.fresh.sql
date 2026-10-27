-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.assert_completed_evidence_v1(
    p_operation_id uuid, p_run_id uuid, p_terminal_evidence_document bytea
)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_terminal jsonb;
    v_invalid boolean;
BEGIN
    IF SESSION_USER NOT IN ('harpy_cluster_admin','harpy_cutover_finalizer') OR
       p_operation_id IS NULL OR p_run_id IS NULL OR p_terminal_evidence_document IS NULL OR
       pg_catalog.octet_length(p_terminal_evidence_document) NOT BETWEEN 2 AND 16777216 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover completion evidence input is invalid';
    END IF;
    v_terminal:=pg_catalog.convert_from(p_terminal_evidence_document,'UTF8')::jsonb;
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    IF v_terminal->>'terminal_state' IS DISTINCT FROM 'completed' OR
       v_terminal->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       v_terminal->>'run_id' IS DISTINCT FROM p_run_id::text THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover completion evidence identity differs';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.cutover_initial_copy_receipts AS receipt
        WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
          AND receipt.manifest_sha256=v_terminal->'initial_copy'->>'manifest_sha256'
          AND receipt.manifest_sha256=v_operation.lease_manifest_sha256
          AND receipt.snapshot_name=v_terminal->'initial_copy'->>'snapshot_name'
          AND receipt.snapshot_lsn=(v_terminal->'initial_copy'->>'snapshot_lsn')::pg_lsn
          AND receipt.relation_count=(v_terminal->'initial_copy'->>'relation_count')::integer
          AND receipt.partition_count=(v_terminal->'initial_copy'->>'partition_count')::integer
          AND receipt.row_count=(v_terminal->'initial_copy'->>'row_count')::bigint
          AND receipt.root_sha256=v_terminal->'initial_copy'->>'root_sha256') OR
       (SELECT pg_catalog.count(*) FROM platform.cutover_operation_relations AS relation
          WHERE relation.operation_id=p_operation_id AND relation.run_id=p_run_id)
            IS DISTINCT FROM (v_terminal->'initial_copy'->>'relation_count')::bigint OR
       (SELECT pg_catalog.count(*) FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
          WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id)
            IS DISTINCT FROM (v_terminal->'initial_copy'->>'partition_count')::bigint OR
       NOT EXISTS (SELECT 1 FROM platform.cutover_operation_relations AS relation
          WHERE relation.operation_id=p_operation_id AND relation.run_id=p_run_id
          HAVING pg_catalog.min(relation.ordinal)=1
             AND pg_catalog.max(relation.ordinal)=pg_catalog.count(*)
             AND pg_catalog.count(*)=(v_terminal->'initial_copy'->>'relation_count')::bigint) OR
       EXISTS ((SELECT relation.ordinal FROM platform.cutover_operation_relations AS relation
          WHERE relation.operation_id=p_operation_id AND relation.run_id=p_run_id)
          EXCEPT (SELECT DISTINCT (item->>'relation_ordinal')::smallint
             FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') AS item)) OR
       EXISTS ((SELECT DISTINCT (item->>'relation_ordinal')::smallint
             FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') AS item)
          EXCEPT (SELECT relation.ordinal FROM platform.cutover_operation_relations AS relation
             WHERE relation.operation_id=p_operation_id AND relation.run_id=p_run_id)) OR
       EXISTS ((SELECT checkpoint_key.relation_ordinal,checkpoint_key.checkpoint_key
          FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
          WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id)
          EXCEPT (SELECT (item->>'relation_ordinal')::smallint,item->>'checkpoint_key'
             FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') AS item)) OR
       EXISTS ((SELECT (item->>'relation_ordinal')::smallint,item->>'checkpoint_key'
             FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') AS item)
          EXCEPT (SELECT checkpoint_key.relation_ordinal,checkpoint_key.checkpoint_key
             FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
             WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover completion checkpoint set differs';
    END IF;
    WITH checkpoint_evidence AS (
        SELECT (item->>'relation_ordinal')::smallint AS relation_ordinal,item->>'checkpoint_key' AS checkpoint_key,
            (item->>'applied_lsn')::pg_lsn AS applied_lsn,(item->>'applied_rows')::bigint AS applied_rows,
            (item->>'batch_count')::integer AS batch_count,item->>'batch_root_sha256' AS batch_root_sha256
        FROM pg_catalog.jsonb_array_elements(v_terminal->'checkpoints') AS item
    ), batch_leaf AS (
        SELECT batch.relation_ordinal,batch.checkpoint_key,batch.batch_ordinal,
          pg_catalog.octet_length(batch.batch_ordinal::text)::text||':'||batch.batch_ordinal::text||
          pg_catalog.octet_length(batch.source_xid::text)::text||':'||batch.source_xid::text||
          pg_catalog.octet_length(batch.source_commit_lsn::text)::text||':'||batch.source_commit_lsn::text||
          '64:'||batch.batch_sha256||pg_catalog.octet_length(batch.applied_count::text)::text||':'||batch.applied_count::text||
          '64:'||batch.key_state_sha256||'64:'||batch.state_sha256 AS leaf,
          batch.applied_count,batch.source_commit_lsn
        FROM platform.cutover_batch_receipts AS batch
        WHERE batch.operation_id=p_operation_id AND batch.run_id=p_run_id
    ), batch_summary AS (
        SELECT leaf.relation_ordinal,leaf.checkpoint_key,pg_catalog.count(*) AS batch_count,
          pg_catalog.sum(leaf.applied_count) AS applied_rows,pg_catalog.max(leaf.source_commit_lsn) AS applied_lsn,
          pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to('cutover-batch-root.v1'||pg_catalog.chr(10)||
            pg_catalog.string_agg(pg_catalog.octet_length(leaf.leaf)::text||':'||leaf.leaf,'' ORDER BY leaf.batch_ordinal),'UTF8')),'hex') AS batch_root_sha256
        FROM batch_leaf AS leaf GROUP BY leaf.relation_ordinal,leaf.checkpoint_key
    ) SELECT EXISTS (SELECT 1 FROM checkpoint_evidence AS terminal
        LEFT JOIN platform.cutover_copy_checkpoints AS checkpoint ON checkpoint.operation_id=p_operation_id
          AND checkpoint.run_id=p_run_id AND checkpoint.relation_ordinal=terminal.relation_ordinal AND checkpoint.checkpoint_key=terminal.checkpoint_key
        LEFT JOIN batch_summary AS batch ON batch.relation_ordinal=terminal.relation_ordinal AND batch.checkpoint_key=terminal.checkpoint_key
        WHERE checkpoint.operation_id IS NULL OR checkpoint.state IS DISTINCT FROM 'complete' OR checkpoint.blocked_reason IS NOT NULL
          OR checkpoint.manifest_sha256 IS DISTINCT FROM v_operation.lease_manifest_sha256
          OR checkpoint.snapshot_name IS DISTINCT FROM v_terminal->'initial_copy'->>'snapshot_name'
          OR checkpoint.snapshot_lsn IS DISTINCT FROM (v_terminal->'initial_copy'->>'snapshot_lsn')::pg_lsn
          OR checkpoint.final_lsn IS DISTINCT FROM (v_terminal->>'final_lsn')::pg_lsn
          OR checkpoint.applied_lsn IS DISTINCT FROM terminal.applied_lsn OR checkpoint.applied_lsn<(v_terminal->>'final_lsn')::pg_lsn
          OR checkpoint.received_lsn<checkpoint.applied_lsn OR checkpoint.applied_rows IS DISTINCT FROM terminal.applied_rows
          OR checkpoint.attempt IS DISTINCT FROM terminal.batch_count OR batch.batch_count IS DISTINCT FROM terminal.batch_count::bigint
          OR batch.applied_rows IS DISTINCT FROM terminal.applied_rows OR batch.applied_lsn IS DISTINCT FROM terminal.applied_lsn
          OR batch.batch_root_sha256 IS DISTINCT FROM terminal.batch_root_sha256)
       OR EXISTS (SELECT 1 FROM platform.cutover_batch_receipts AS batch
          JOIN platform.cutover_copy_checkpoints AS checkpoint USING (operation_id,run_id,relation_ordinal,checkpoint_key)
          WHERE batch.operation_id=p_operation_id AND batch.run_id=p_run_id AND batch.applied_count=0
            AND (batch.source_commit_lsn IS DISTINCT FROM checkpoint.final_lsn OR batch.batch_ordinal IS DISTINCT FROM
              (SELECT pg_catalog.max(candidate.batch_ordinal) FROM platform.cutover_batch_receipts AS candidate
               WHERE candidate.operation_id=batch.operation_id AND candidate.run_id=batch.run_id
                 AND candidate.relation_ordinal=batch.relation_ordinal AND candidate.checkpoint_key=batch.checkpoint_key)))
       OR EXISTS (SELECT 1 FROM platform.cutover_batch_receipts AS batch
          WHERE batch.operation_id=p_operation_id AND batch.run_id=p_run_id AND batch.applied_count=0
          GROUP BY batch.relation_ordinal,batch.checkpoint_key HAVING pg_catalog.count(*)>1) INTO v_invalid;
    IF v_invalid THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover completion batch evidence differs';
    END IF;
    WITH change_leaf AS (
        SELECT receipt.source_xid,receipt.source_commit_lsn,receipt.relation_ordinal,receipt.change_ordinal,receipt.checkpoint_key,
          pg_catalog.octet_length(receipt.relation_ordinal::text)::text||':'||receipt.relation_ordinal::text||
          pg_catalog.octet_length(receipt.change_ordinal::text)::text||':'||receipt.change_ordinal::text||
          pg_catalog.octet_length(receipt.checkpoint_key)::text||':'||receipt.checkpoint_key||
          pg_catalog.octet_length(receipt.mutation_kind)::text||':'||receipt.mutation_kind||'64:'||receipt.mutation_sha256 AS leaf
        FROM platform.applied_change_receipts AS receipt WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
    ), change_group AS (
        SELECT leaf.source_xid,pg_catalog.min(leaf.source_commit_lsn) AS source_commit_lsn,
          pg_catalog.count(*) AS change_count,pg_catalog.count(DISTINCT leaf.source_commit_lsn) AS commit_count,
          pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to('cutover-journal-changes-root.v1'||pg_catalog.chr(10)||
            pg_catalog.string_agg(pg_catalog.octet_length(leaf.leaf)::text||':'||leaf.leaf,'' ORDER BY leaf.relation_ordinal,leaf.change_ordinal,leaf.checkpoint_key),'UTF8')),'hex') AS changes_sha256
        FROM change_leaf AS leaf GROUP BY leaf.source_xid
    ), coverage_leaf AS (
        SELECT change.source_xid,change.source_commit_lsn,change.change_count,change.commit_count,
          pg_catalog.octet_length(change.source_xid::text)::text||':'||change.source_xid::text||
          pg_catalog.octet_length(change.source_commit_lsn::text)::text||':'||change.source_commit_lsn::text||
          pg_catalog.octet_length(change.change_count::text)::text||':'||change.change_count::text||'64:'||change.changes_sha256 AS leaf
        FROM change_group AS change
    ), coverage AS (
        SELECT pg_catalog.count(*) AS transaction_count,COALESCE(pg_catalog.sum(leaf.change_count),0) AS change_count,
          pg_catalog.max(leaf.commit_count) AS max_commit_count,
          pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to('cutover-journal-coverage-root.v1'||pg_catalog.chr(10)||
            COALESCE(pg_catalog.string_agg(pg_catalog.octet_length(leaf.leaf)::text||':'||leaf.leaf,'' ORDER BY leaf.source_xid),''),'UTF8')),'hex') AS root_sha256
        FROM coverage_leaf AS leaf
    ), delta_batch_group AS (
        SELECT batch.source_xid,pg_catalog.min(batch.source_commit_lsn) AS min_lsn,
          pg_catalog.max(batch.source_commit_lsn) AS max_lsn,pg_catalog.sum(batch.applied_count) AS applied_count
        FROM platform.cutover_batch_receipts AS batch WHERE batch.operation_id=p_operation_id AND batch.run_id=p_run_id
          AND EXISTS (SELECT 1 FROM platform.cutover_change_journal AS journal
            WHERE journal.operation_id=p_operation_id AND journal.run_id=p_run_id AND journal.source_xid=batch.source_xid)
        GROUP BY batch.source_xid
    ) SELECT EXISTS (SELECT 1 FROM platform.cutover_change_journal AS journal
          WHERE journal.operation_id=p_operation_id AND journal.run_id=p_run_id AND
            (journal.scope_kind IS DISTINCT FROM v_operation.scope_kind OR journal.scope_id IS DISTINCT FROM v_operation.scope_id
             OR journal.domain IS DISTINCT FROM v_operation.domain OR journal.cutover_generation IS DISTINCT FROM v_operation.cutover_generation))
       OR EXISTS ((SELECT journal.source_xid FROM platform.cutover_change_journal AS journal WHERE journal.operation_id=p_operation_id AND journal.run_id=p_run_id)
          EXCEPT (SELECT receipt.source_xid FROM platform.applied_change_receipts AS receipt WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id))
       OR EXISTS ((SELECT receipt.source_xid FROM platform.applied_change_receipts AS receipt WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id)
          EXCEPT (SELECT journal.source_xid FROM platform.cutover_change_journal AS journal WHERE journal.operation_id=p_operation_id AND journal.run_id=p_run_id))
       OR EXISTS (SELECT 1 FROM platform.applied_change_receipts AS receipt WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
          AND (receipt.manifest_sha256 IS DISTINCT FROM v_operation.lease_manifest_sha256 OR receipt.source_commit_lsn>(v_terminal->>'final_lsn')::pg_lsn))
       OR EXISTS (SELECT 1 FROM platform.cutover_batch_receipts AS batch
          WHERE batch.operation_id=p_operation_id AND batch.run_id=p_run_id AND batch.applied_count>0
            AND NOT EXISTS (SELECT 1 FROM platform.cutover_change_journal AS journal
              WHERE journal.operation_id=p_operation_id AND journal.run_id=p_run_id AND journal.source_xid=batch.source_xid))
       OR EXISTS (SELECT 1 FROM coverage_leaf AS applied FULL JOIN delta_batch_group AS batch USING (source_xid)
          WHERE applied.change_count IS DISTINCT FROM batch.applied_count OR applied.commit_count IS DISTINCT FROM 1::bigint
             OR applied.source_commit_lsn IS DISTINCT FROM batch.min_lsn OR batch.min_lsn IS DISTINCT FROM batch.max_lsn)
       OR NOT EXISTS (SELECT 1 FROM coverage WHERE transaction_count=(v_terminal->>'journal_transaction_count')::bigint
          AND change_count=(v_terminal->>'applied_change_count')::bigint AND COALESCE(max_commit_count,1)=1
          AND root_sha256=v_terminal->>'journal_coverage_root_sha256') INTO v_invalid;
    IF v_invalid THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover journal coverage differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.assert_completed_evidence_v1(uuid,uuid,bytea) FROM PUBLIC;
