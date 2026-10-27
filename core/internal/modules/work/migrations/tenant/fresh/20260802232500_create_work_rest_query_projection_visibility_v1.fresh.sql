-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_projection_principals_valid_v1(p_principals text[], p_visible boolean)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER
SET search_path FROM CURRENT AS $function$
BEGIN
    RETURN p_principals IS NOT NULL
       AND pg_catalog.cardinality(p_principals) <= 100
       AND (NOT p_visible OR pg_catalog.cardinality(p_principals) BETWEEN 1 AND 100)
       AND pg_catalog.array_position(p_principals, NULL) IS NULL
       AND NOT EXISTS (SELECT 1 FROM pg_catalog.unnest(p_principals) AS principal(value)
           WHERE principal.value !~ '^(USER|ROLE|ORGANIZATION):[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
       AND pg_catalog.cardinality(p_principals) =
           (SELECT pg_catalog.count(DISTINCT principal.value)
            FROM pg_catalog.unnest(p_principals) AS principal(value));
END;
$function$;
REVOKE ALL ON FUNCTION work_projection_principals_valid_v1(text[], boolean) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION work_projection_principals_valid_v1(text[], boolean) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION work_projection_principals_valid_v1(text[], boolean) TO forgecrud_adopter;

ALTER TABLE work_inbox_items ADD COLUMN visible_to_principals text[] NOT NULL DEFAULT ARRAY[]::text[];
ALTER TABLE work_calendar_items ADD COLUMN visible_to_principals text[] NOT NULL DEFAULT ARRAY[]::text[];
UPDATE work_inbox_items SET state='STALE' WHERE state='VISIBLE' AND pg_catalog.cardinality(visible_to_principals)=0;
UPDATE work_calendar_items SET state='STALE' WHERE state='VISIBLE' AND pg_catalog.cardinality(visible_to_principals)=0;
ALTER TABLE work_inbox_items ADD CONSTRAINT work_inbox_visibility_principals_check
    CHECK (work_projection_principals_valid_v1(visible_to_principals, state='VISIBLE'));
ALTER TABLE work_calendar_items ADD CONSTRAINT work_calendar_visibility_principals_check
    CHECK (work_projection_principals_valid_v1(visible_to_principals, state='VISIBLE'));
DROP FUNCTION work_v1_adoption_document_valid_v1(text,jsonb);
CREATE FUNCTION work_v1_adoption_document_valid_v1(p_target_kind text, p_record jsonb)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER
SET search_path FROM CURRENT AS $function$
DECLARE v_keys text; v_expected text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' OR p_record='{}'::jsonb THEN RETURN false; END IF;
    SELECT pg_catalog.string_agg(key,',' ORDER BY key) INTO v_keys
      FROM pg_catalog.jsonb_object_keys(p_record) AS key;
    v_expected := CASE p_target_kind
      WHEN 'BOARD' THEN 'ArchivedAt,CreatedAt,CreatorID,Description,Lifecycle,Name,UpdatedAt'
      WHEN 'COLUMN' THEN 'BoardID,CreatedAt,Name,Position,Status,UpdatedAt'
      WHEN 'MEMBERSHIP' THEN 'BoardID,CreatedAt,PrincipalID,PrincipalKind,Role,UpdatedAt'
      WHEN 'TASK' THEN 'BoardID,CancelledAt,ColumnID,CompletedAt,CreatedAt,DeletedAt,Description,DueAt,Kind,Position,Status,Title,UpdatedAt'
      WHEN 'ASSIGNEE' THEN 'PrincipalID,PrincipalKind,TaskID' WHEN 'LABEL' THEN 'BoardID,Color,Name'
      WHEN 'TASK_LABEL' THEN 'LabelID,TaskID' WHEN 'CHECKLIST' THEN 'Completed,Position,TaskID,Text'
      WHEN 'COMMENT' THEN 'AuthorID,BoardID,Body,CreatedAt,RedactedAt,RedactedBy,RedactionKind,RedactionReason,State,TaskID'
      WHEN 'CALENDAR' THEN 'Audience,BoardID,CreatedAt,DeletedAt,Description,EndDate,EndsAt,OrganizationID,OwnerID,StartDate,StartsAt,Timezone,Title,UpdatedAt'
      WHEN 'RECURRENCE' THEN 'DTStart,EventID,HorizonEnd,RRULE,TZDBVersion,Timezone'
      WHEN 'FORM_LINK' THEN 'FormID,PermissionRevision,ReferenceEpoch,ResourceID,ResourceKind,VersionID'
      WHEN 'WORKFLOW_PROJECTION' THEN 'DueAt,PermissionRevision,Revision,SourceID,Summary,VersionID,VisibleToPrincipals'
      ELSE NULL END;
    IF v_keys IS DISTINCT FROM v_expected THEN RETURN false; END IF;
    IF p_target_kind IS DISTINCT FROM 'WORKFLOW_PROJECTION' THEN RETURN true; END IF;
    IF pg_catalog.jsonb_typeof(p_record->'VisibleToPrincipals') IS DISTINCT FROM 'array' THEN RETURN false; END IF;
    RETURN work_projection_principals_valid_v1(ARRAY(SELECT value
      FROM pg_catalog.jsonb_array_elements_text(p_record->'VisibleToPrincipals') AS principal(value)
      ORDER BY value),true);
EXCEPTION WHEN invalid_parameter_value OR data_exception THEN RETURN false;
END;
$function$;
REVOKE ALL ON FUNCTION work_v1_adoption_document_valid_v1(text,jsonb) FROM PUBLIC;

CREATE FUNCTION work_v1_adoption_inbox_visibility_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER
SET search_path FROM CURRENT AS $function$
DECLARE v_matches bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR NEW.source_owner IS DISTINCT FROM 'WORKFLOW' THEN RETURN NEW; END IF;
    SELECT pg_catalog.array_agg(principal.value ORDER BY principal.value),
           pg_catalog.count(DISTINCT stage.run_id) INTO NEW.visible_to_principals,v_matches
      FROM work_v1_adoption_stage AS stage JOIN work_adoption_runs AS run ON run.run_id=stage.run_id
      CROSS JOIN pg_catalog.jsonb_array_elements_text(stage.target_record->'VisibleToPrincipals') AS principal(value)
     WHERE stage.target_kind='WORKFLOW_PROJECTION' AND stage.target_id=NEW.id AND NOT stage.removed
       AND run.state IN ('COPYING','RECONCILING') AND stage.target_record->>'SourceID'=NEW.source_id::text
       AND stage.target_record->>'VersionID'=NEW.source_version_id::text
       AND (stage.target_record->>'Revision')::bigint=NEW.source_revision
       AND (stage.target_record->>'PermissionRevision')::bigint=NEW.permission_revision
       AND stage.target_record->>'Summary'=NEW.summary
       AND (stage.target_record->>'DueAt')::timestamptz IS NOT DISTINCT FROM NEW.due_at;
    IF v_matches IS DISTINCT FROM 1::bigint OR NOT work_projection_principals_valid_v1(NEW.visible_to_principals,true) THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='Work V1 workflow projection visibility denied'; END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION work_v1_adoption_inbox_visibility_v1() FROM PUBLIC;
CREATE TRIGGER work_v1_adoption_inbox_visibility_trigger_v1 BEFORE INSERT OR UPDATE ON work_inbox_items
    FOR EACH ROW EXECUTE FUNCTION work_v1_adoption_inbox_visibility_v1();

CREATE VIEW work_query_board_visibility_v1 AS
SELECT membership.board_id,
       pg_catalog.array_agg(DISTINCT membership.principal_kind||':'||membership.principal_id::text
           ORDER BY membership.principal_kind||':'||membership.principal_id::text) AS visibility_principals
FROM work_board_memberships AS membership WHERE membership.active GROUP BY membership.board_id;

CREATE VIEW work_board_query_v1 AS
SELECT board.id,board.name,pg_catalog.normalize(pg_catalog.casefold(
       pg_catalog.normalize(board.name,'NFC') COLLATE pg_catalog.pg_unicode_fast),'NFC') AS name_folded,
       board.description,board.creator_user_id AS creator_id,board.lifecycle_state AS lifecycle,board.row_version,
       (SELECT pg_catalog.count(*) FROM work_tasks AS task WHERE task.board_id=board.id AND task.deleted_at IS NULL) AS task_count,
       (SELECT pg_catalog.count(*) FROM work_board_memberships AS member_row WHERE member_row.board_id=board.id AND member_row.active) AS member_count,
       board.created_at,board.updated_at,visibility.visibility_principals
FROM work_boards AS board JOIN work_query_board_visibility_v1 AS visibility ON visibility.board_id=board.id
WHERE SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime' AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW work_task_query_v1 AS
SELECT task.id,task.board_id,task.column_id,task.title,pg_catalog.normalize(pg_catalog.casefold(
       pg_catalog.normalize(task.title,'NFC') COLLATE pg_catalog.pg_unicode_fast),'NFC') AS title_folded,
       task.description,task.task_kind AS kind,task.status_category AS status,task.priority,
       task.creator_user_id AS creator_id,task.position,task.row_version,task.starts_at,task.due_at,
       task.created_at,task.updated_at,visibility.visibility_principals
FROM work_tasks AS task JOIN work_query_board_visibility_v1 AS visibility ON visibility.board_id=task.board_id
WHERE task.deleted_at IS NULL AND SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW work_comment_query_v1 AS
SELECT comment.id,comment.board_id,comment.task_id,comment.author_id,comment.body,
       pg_catalog.normalize(pg_catalog.casefold(pg_catalog.normalize(comment.body,'NFC')
           COLLATE pg_catalog.pg_unicode_fast),'NFC') AS body_folded,
       comment.state,comment.redaction_kind,comment.row_version,comment.created_at,comment.redacted_at,
       visibility.visibility_principals
FROM work_comments AS comment JOIN work_query_board_visibility_v1 AS visibility ON visibility.board_id=comment.board_id
WHERE SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime' AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW work_calendar_event_query_v1 AS
SELECT event.id,event.owner_id,event.audience,event.board_id,event.organization_id,event.title,
       pg_catalog.normalize(pg_catalog.casefold(pg_catalog.normalize(event.title,'NFC')
           COLLATE pg_catalog.pg_unicode_fast),'NFC') AS title_folded,
       event.description,event.color,event.starts_at,event.ends_at,event.start_date,event.end_date,
       event.timezone_name AS timezone,event.generated,event.row_version,event.created_at,event.updated_at,
       CASE event.audience WHEN 'PRIVATE' THEN ARRAY['USER:'||event.owner_id::text]::text[]
           WHEN 'ORGANIZATION' THEN ARRAY['ORGANIZATION:'||event.organization_id::text]::text[]
           ELSE COALESCE(visibility.visibility_principals,ARRAY[]::text[]) END AS visibility_principals
FROM work_calendar_events AS event LEFT JOIN work_query_board_visibility_v1 AS visibility ON visibility.board_id=event.board_id
WHERE event.deleted_at IS NULL AND SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW work_inbox_query_v1 AS
SELECT task.id,'WORK'::text AS source_owner,'TASK'::text AS source_kind,task.id AS source_id,
       task.id AS source_version_id,task.row_version AS source_revision,0::bigint AS permission_revision,
       task.title AS summary,pg_catalog.normalize(pg_catalog.casefold(pg_catalog.normalize(task.title,'NFC')
           COLLATE pg_catalog.pg_unicode_fast),'NFC') AS summary_folded,
       task.due_at,task.board_id,task.status_category AS status,task.priority,task.updated_at,
       assignee.visibility_principals,board_visibility.visibility_principals AS context_principals
FROM work_tasks AS task JOIN work_query_board_visibility_v1 AS board_visibility ON board_visibility.board_id=task.board_id
JOIN LATERAL (SELECT pg_catalog.array_agg(assignee_row.principal_kind||':'||assignee_row.principal_id::text
        ORDER BY assignee_row.principal_kind,assignee_row.principal_id) AS visibility_principals
    FROM work_task_assignees AS assignee_row WHERE assignee_row.task_id=task.id AND assignee_row.active) AS assignee
    ON pg_catalog.cardinality(assignee.visibility_principals)>0
WHERE task.deleted_at IS NULL AND SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime'
UNION ALL
SELECT item.id,item.source_owner,item.source_kind,item.source_id,item.source_version_id,item.source_revision,
       item.permission_revision,item.summary,pg_catalog.normalize(pg_catalog.casefold(
       pg_catalog.normalize(item.summary,'NFC') COLLATE pg_catalog.pg_unicode_fast),'NFC') AS summary_folded,
       item.due_at,NULL::uuid,NULL::text,NULL::text,item.updated_at,item.visible_to_principals,item.visible_to_principals
FROM work_inbox_items AS item WHERE item.state='VISIBLE'
  AND SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime' AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW work_calendar_query_v1 AS
SELECT inbox.*,'TASK_DEADLINE'::text AS item_kind,inbox.due_at AS starts_at,NULL::timestamptz AS ends_at,
       NULL::date AS start_date,NULL::date AS end_date FROM work_inbox_query_v1 AS inbox
WHERE inbox.source_owner='WORK' AND inbox.due_at IS NOT NULL
UNION ALL
SELECT event.id,'WORK'::text,'CALENDAR_EVENT'::text,event.id,event.id,event.row_version,0::bigint,
       event.title,event.title_folded,NULL::timestamptz,event.board_id,NULL::text,NULL::text,event.updated_at,
       event.visibility_principals,event.visibility_principals,'CALENDAR_EVENT'::text,event.starts_at,event.ends_at,
       event.start_date,event.end_date FROM work_calendar_event_query_v1 AS event
UNION ALL
SELECT item.id,item.source_owner,item.source_kind,item.source_id,item.source_version_id,item.source_revision,
       item.permission_revision,item.summary,pg_catalog.normalize(pg_catalog.casefold(pg_catalog.normalize(item.summary,'NFC')
       COLLATE pg_catalog.pg_unicode_fast),'NFC'),NULL::timestamptz,NULL::uuid,NULL::text,NULL::text,item.updated_at,
       item.visible_to_principals,item.visible_to_principals,'PROJECTION'::text,item.starts_at,item.ends_at,
       item.start_date,item.end_date FROM work_calendar_items AS item WHERE item.state='VISIBLE'
  AND SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime' AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

GRANT SELECT ON work_board_query_v1,work_task_query_v1,work_comment_query_v1,
    work_calendar_event_query_v1,work_inbox_query_v1,work_calendar_query_v1 TO forgecrud_runtime;
GRANT INSERT (creator_user_id,priority,starts_at) ON work_tasks TO forgecrud_runtime;
GRANT UPDATE (priority,starts_at) ON work_tasks TO forgecrud_runtime;
GRANT INSERT (color) ON work_board_columns TO forgecrud_runtime;
GRANT UPDATE (color) ON work_board_columns TO forgecrud_runtime;
GRANT INSERT (color) ON work_calendar_events TO forgecrud_runtime;
GRANT UPDATE (color) ON work_calendar_events TO forgecrud_runtime;
GRANT INSERT (visible_to_principals) ON work_inbox_items TO forgecrud_runtime;
GRANT UPDATE (visible_to_principals) ON work_inbox_items TO forgecrud_runtime;
GRANT INSERT (visible_to_principals) ON work_calendar_items TO forgecrud_runtime;
GRANT UPDATE (visible_to_principals) ON work_calendar_items TO forgecrud_runtime;
ALTER TABLE work_inbox_items ALTER COLUMN visible_to_principals DROP DEFAULT;
ALTER TABLE work_calendar_items ALTER COLUMN visible_to_principals DROP DEFAULT;
