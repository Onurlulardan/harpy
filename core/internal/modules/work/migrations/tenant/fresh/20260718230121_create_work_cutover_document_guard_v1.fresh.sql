-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_document_valid_v1(
    p_target_kind text,
    p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_keys text;
    v_expected text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' OR
       p_record = '{}'::jsonb THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.string_agg(key, ',' ORDER BY key)
    INTO v_keys
    FROM pg_catalog.jsonb_object_keys(p_record) AS key;
    v_expected := CASE p_target_kind
        WHEN 'BOARD' THEN
            'ArchivedAt,CreatedAt,CreatorID,Description,Lifecycle,Name,UpdatedAt'
        WHEN 'COLUMN' THEN 'BoardID,CreatedAt,Name,Position,Status,UpdatedAt'
        WHEN 'MEMBERSHIP' THEN
            'BoardID,CreatedAt,PrincipalID,PrincipalKind,Role,UpdatedAt'
        WHEN 'TASK' THEN
            'BoardID,CancelledAt,ColumnID,CompletedAt,CreatedAt,DeletedAt,' ||
            'Description,DueAt,Kind,Position,Status,Title,UpdatedAt'
        WHEN 'ASSIGNEE' THEN 'PrincipalID,PrincipalKind,TaskID'
        WHEN 'LABEL' THEN 'BoardID,Color,Name'
        WHEN 'TASK_LABEL' THEN 'LabelID,TaskID'
        WHEN 'CHECKLIST' THEN 'Completed,Position,TaskID,Text'
        WHEN 'COMMENT' THEN
            'AuthorID,BoardID,Body,CreatedAt,RedactedAt,RedactedBy,' ||
            'RedactionKind,RedactionReason,State,TaskID'
        WHEN 'CALENDAR' THEN
            'Audience,BoardID,CreatedAt,DeletedAt,Description,EndDate,EndsAt,' ||
            'OrganizationID,OwnerID,StartDate,StartsAt,Timezone,Title,UpdatedAt'
        WHEN 'RECURRENCE' THEN 'DTStart,EventID,HorizonEnd,RRULE,TZDBVersion,Timezone'
        WHEN 'FORM_LINK' THEN
            'FormID,PermissionRevision,ReferenceEpoch,ResourceID,ResourceKind,VersionID'
        WHEN 'WORKFLOW_PROJECTION' THEN
            'DueAt,PermissionRevision,Revision,SourceID,Summary,VersionID'
        ELSE NULL
    END;
    RETURN v_keys IS NOT DISTINCT FROM v_expected;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_document_valid_v1(
    text, jsonb
) FROM PUBLIC;
