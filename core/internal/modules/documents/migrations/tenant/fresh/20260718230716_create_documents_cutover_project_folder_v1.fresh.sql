-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_project_folder_v1(p_record jsonb)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_row documents_folders%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 folder projection principal denied';
    END IF;
    SELECT * INTO STRICT v_row
    FROM pg_catalog.jsonb_populate_record(NULL::documents_folders,p_record);
    INSERT INTO documents_folders (
        id,parent_id,owner_id,display_name,lifecycle_state,row_version,
        created_at,updated_at,trashed_at
    ) VALUES (
        v_row.id,v_row.parent_id,v_row.owner_id,v_row.display_name,
        v_row.lifecycle_state,v_row.row_version,v_row.created_at,
        v_row.updated_at,v_row.trashed_at
    ) ON CONFLICT (id) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM documents_folders AS target
        WHERE target.id=v_row.id AND pg_catalog.to_jsonb(target)=p_record
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 folder projection differs';
    END IF;
    RETURN v_row.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_project_folder_v1(jsonb)
FROM PUBLIC;
