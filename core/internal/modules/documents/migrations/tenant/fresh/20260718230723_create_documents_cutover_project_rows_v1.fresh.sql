-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_project_rows_v1(
    p_operation_id uuid,p_cutover_run_id uuid,p_scope_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage documents_v1_adoption_stage%ROWTYPE;
    v_adoption_run_id uuid;
    v_final_lsn pg_lsn;
    v_binding_root text;
    v_effect_root text;
    v_row_root text;
    v_folder_count bigint;
    v_item_count bigint;
    v_version_count bigint;
BEGIN
    SELECT ready.adoption_run_id,ready.final_lsn,
        ready.binding_root_sha256,ready.effect_root_sha256
    INTO STRICT v_adoption_run_id,v_final_lsn,v_binding_root,v_effect_root
    FROM documents_v1_adoption_projection_ready_v1(
        p_operation_id,p_cutover_run_id,p_scope_id
    ) AS ready;
    SET CONSTRAINTS ALL DEFERRED;
    FOR v_stage IN
        WITH RECURSIVE folder_tree AS (
            SELECT stage.*,0 AS depth
            FROM documents_v1_adoption_stage AS stage
            WHERE stage.operation_id=p_operation_id
              AND stage.cutover_run_id=p_cutover_run_id
              AND stage.target_kind='folder' AND NOT stage.removed
              AND stage.target_record->>'parent_id' IS NULL
            UNION ALL
            SELECT child.*,parent.depth+1
            FROM documents_v1_adoption_stage AS child
            JOIN folder_tree AS parent
              ON (child.target_record->>'parent_id')::uuid=parent.target_id
            WHERE child.operation_id=p_operation_id
              AND child.cutover_run_id=p_cutover_run_id
              AND child.target_kind='folder' AND NOT child.removed
              AND parent.depth<1024
        )
        SELECT tree.* FROM folder_tree AS tree
        ORDER BY tree.depth,tree.target_id
    LOOP
        PERFORM documents_v1_adoption_project_folder_v1(v_stage.target_record);
    END LOOP;
    FOR v_stage IN
        SELECT stage.* FROM documents_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id
          AND stage.target_kind='item' AND NOT stage.removed
        ORDER BY stage.target_id
    LOOP
        PERFORM documents_v1_adoption_project_item_v1(v_stage.target_record);
    END LOOP;
    FOR v_stage IN
        SELECT stage.* FROM documents_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id
          AND stage.target_kind='version' AND NOT stage.removed
        ORDER BY stage.target_id
    LOOP
        PERFORM documents_v1_adoption_project_version_v1(
            v_adoption_run_id,v_stage.target_record
        );
    END LOOP;
    SELECT pg_catalog.count(*) INTO v_folder_count
    FROM documents_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.cutover_run_id=p_cutover_run_id
      AND stage.target_kind='folder' AND NOT stage.removed;
    SELECT pg_catalog.count(*) INTO v_item_count
    FROM documents_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.cutover_run_id=p_cutover_run_id
      AND stage.target_kind='item' AND NOT stage.removed;
    SELECT pg_catalog.count(*) INTO v_version_count
    FROM documents_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.cutover_run_id=p_cutover_run_id
      AND stage.target_kind='version' AND NOT stage.removed;
    IF (SELECT pg_catalog.count(*) FROM documents_folders)<>v_folder_count OR
       (SELECT pg_catalog.count(*) FROM documents_items)<>v_item_count OR
       (SELECT pg_catalog.count(*) FROM documents_versions)<>v_version_count OR
       EXISTS (
           SELECT 1 FROM documents_v1_adoption_stage AS stage
           WHERE stage.operation_id=p_operation_id
             AND stage.cutover_run_id=p_cutover_run_id
             AND stage.target_kind IN ('folder','item') AND NOT stage.removed
             AND NOT CASE stage.target_kind
                 WHEN 'folder' THEN EXISTS (
                     SELECT 1 FROM documents_folders AS target
                     WHERE target.id=stage.target_id
                       AND pg_catalog.to_jsonb(target)=stage.target_record
                 )
                 WHEN 'item' THEN EXISTS (
                     SELECT 1 FROM documents_items AS target
                     WHERE target.id=stage.target_id
                       AND pg_catalog.to_jsonb(target)=stage.target_record
                 )
             END
       ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 projected rows differ';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        COALESCE(pg_catalog.string_agg(stage.target_kind||':'||stage.target_identity||
            ':'||pg_catalog.encode(stage.target_sha256,'hex'),',' ORDER BY
            stage.target_kind,stage.target_identity),''),'UTF8'
    )),'hex') INTO v_row_root
    FROM documents_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.cutover_run_id=p_cutover_run_id AND NOT stage.removed;
    RETURN v_row_root;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_project_rows_v1(uuid,uuid,uuid)
FROM PUBLIC;
