-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_mapping_ready_v1(
    p_operation_id uuid,p_cutover_run_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_adoption_run_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 mapping principal denied';
    END IF;
    SELECT (stage.target_record->>'adoption_run_id')::uuid
    INTO STRICT v_adoption_run_id
    FROM documents_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.cutover_run_id=p_cutover_run_id
      AND stage.target_kind='inventory' AND NOT stage.removed;
    IF (SELECT pg_catalog.count(*) FROM documents_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id
          AND stage.target_kind='inventory' AND NOT stage.removed)<>1 OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id AND NOT stage.removed
        GROUP BY stage.source_kind,stage.source_identity
        HAVING (stage.source_kind IN ('folder','item','inventory') AND
                (pg_catalog.count(*)<>1 OR
                 pg_catalog.count(DISTINCT stage.target_kind)<>1)) OR
               (stage.source_kind='version' AND
                (pg_catalog.count(*)<>2 OR
                 pg_catalog.count(DISTINCT stage.target_kind)<>2 OR
                 pg_catalog.count(DISTINCT stage.target_id)<>1))
    ) OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id AND NOT stage.removed
          AND stage.source_kind='version'
        GROUP BY stage.source_identity
        HAVING pg_catalog.bool_or(stage.target_kind='version') IS DISTINCT FROM true OR
               pg_catalog.bool_or(stage.target_kind='checkpoint') IS DISTINCT FROM true
    ) OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS folder
        WHERE folder.operation_id=p_operation_id
          AND folder.cutover_run_id=p_cutover_run_id
          AND folder.target_kind='folder' AND NOT folder.removed
          AND folder.target_record->>'parent_id' IS NOT NULL
          AND NOT EXISTS (
              SELECT 1 FROM documents_v1_adoption_stage AS parent
              WHERE parent.operation_id=folder.operation_id
                AND parent.cutover_run_id=folder.cutover_run_id
                AND parent.target_kind='folder' AND NOT parent.removed
                AND parent.target_id=(folder.target_record->>'parent_id')::uuid
          )
    ) OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS item
        WHERE item.operation_id=p_operation_id
          AND item.cutover_run_id=p_cutover_run_id
          AND item.target_kind='item' AND NOT item.removed
          AND NOT EXISTS (
              SELECT 1 FROM documents_v1_adoption_stage AS folder
              WHERE folder.operation_id=item.operation_id
                AND folder.cutover_run_id=item.cutover_run_id
                AND folder.target_kind='folder' AND NOT folder.removed
                AND folder.target_id=(item.target_record->>'folder_id')::uuid
          )
    ) OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS version
        WHERE version.operation_id=p_operation_id
          AND version.cutover_run_id=p_cutover_run_id
          AND version.target_kind='version' AND NOT version.removed
          AND NOT EXISTS (
              SELECT 1 FROM documents_v1_adoption_stage AS item
              WHERE item.operation_id=version.operation_id
                AND item.cutover_run_id=version.cutover_run_id
                AND item.target_kind='item' AND NOT item.removed
                AND item.target_id=(version.target_record->>'item_id')::uuid
          )
    ) OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS item
        WHERE item.operation_id=p_operation_id
          AND item.cutover_run_id=p_cutover_run_id
          AND item.target_kind='item' AND NOT item.removed
          AND item.target_record->>'current_version_id' IS NOT NULL
          AND NOT EXISTS (
              SELECT 1 FROM documents_v1_adoption_stage AS version
              WHERE version.operation_id=item.operation_id
                AND version.cutover_run_id=item.cutover_run_id
                AND version.target_kind='version' AND NOT version.removed
                AND version.target_id=(item.target_record->>'current_version_id')::uuid
                AND (version.target_record->>'item_id')::uuid=item.target_id
          )
    ) OR EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS checkpoint
        WHERE checkpoint.operation_id=p_operation_id
          AND checkpoint.cutover_run_id=p_cutover_run_id
          AND checkpoint.target_kind='checkpoint' AND NOT checkpoint.removed
          AND (checkpoint.target_record->>'adoption_run_id')::uuid
              IS DISTINCT FROM v_adoption_run_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 source-target mapping is incomplete';
    END IF;
    RETURN v_adoption_run_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_mapping_ready_v1(uuid,uuid)
FROM PUBLIC;
