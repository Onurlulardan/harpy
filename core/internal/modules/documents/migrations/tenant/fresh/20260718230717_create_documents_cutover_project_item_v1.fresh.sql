-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_project_item_v1(p_record jsonb)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_row documents_items%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 item projection principal denied';
    END IF;
    SELECT * INTO STRICT v_row
    FROM pg_catalog.jsonb_populate_record(NULL::documents_items,p_record);
    INSERT INTO documents_items (
        id,folder_id,owner_id,current_version_id,lifecycle_state,resume_state,
        row_version,trashed_at,retention_ends_at,purge_requested_at,
        purge_starts_at,purge_reason,purged_at,tombstone_until,
        purge_evidence_sha256,created_at,updated_at,description,tags
    ) VALUES (
        v_row.id,v_row.folder_id,v_row.owner_id,v_row.current_version_id,
        v_row.lifecycle_state,v_row.resume_state,v_row.row_version,v_row.trashed_at,
        v_row.retention_ends_at,v_row.purge_requested_at,v_row.purge_starts_at,
        v_row.purge_reason,v_row.purged_at,v_row.tombstone_until,
        v_row.purge_evidence_sha256,v_row.created_at,v_row.updated_at,
        v_row.description,v_row.tags
    ) ON CONFLICT (id) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM documents_items AS target
        WHERE target.id=v_row.id AND pg_catalog.to_jsonb(target)=p_record
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 item projection differs';
    END IF;
    RETURN v_row.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_project_item_v1(jsonb)
FROM PUBLIC;
