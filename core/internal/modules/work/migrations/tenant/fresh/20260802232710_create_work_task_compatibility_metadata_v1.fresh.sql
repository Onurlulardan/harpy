-- +goose Up
CREATE TABLE work_task_compatibility_metadata (
    task_id uuid PRIMARY KEY,
    is_recurring boolean NOT NULL DEFAULT false,
    recurrence_rule varchar(50) NOT NULL DEFAULT '',
    recurrence_end timestamptz,
    form_filter jsonb,
    created_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT transaction_timestamp(),
    CONSTRAINT work_task_compatibility_task_fk FOREIGN KEY (task_id)
        REFERENCES work_tasks(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_task_compatibility_rule_size_check
        CHECK (pg_catalog.char_length(recurrence_rule) <= 50),
    CONSTRAINT work_task_compatibility_filter_check
        CHECK (form_filter IS NULL
            OR pg_catalog.octet_length(form_filter::text) <= 128000)
);

GRANT SELECT ON work_task_compatibility_metadata TO forgecrud_runtime;
GRANT INSERT (task_id,is_recurring,recurrence_rule,recurrence_end,form_filter,created_at,updated_at)
    ON work_task_compatibility_metadata TO forgecrud_runtime;
GRANT UPDATE (is_recurring,recurrence_rule,recurrence_end,form_filter,updated_at)
    ON work_task_compatibility_metadata TO forgecrud_runtime;

DROP FUNCTION work_v1_adoption_document_valid_v1(text,jsonb);
-- +goose StatementBegin
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
      WHEN 'TASK' THEN 'BoardID,CancelledAt,ColumnID,CompletedAt,CreatedAt,CreatorID,DeletedAt,Description,DueAt,FormFilter,IsRecurring,Kind,Position,Priority,RecurrenceEnd,RecurrenceRule,StartsAt,Status,Title,UpdatedAt'
      WHEN 'ASSIGNEE' THEN 'PrincipalID,PrincipalKind,TaskID' WHEN 'LABEL' THEN 'BoardID,Color,Name'
      WHEN 'TASK_LABEL' THEN 'LabelID,TaskID' WHEN 'CHECKLIST' THEN 'Completed,Position,TaskID,Text'
      WHEN 'COMMENT' THEN 'AuthorID,BoardID,Body,CreatedAt,RedactedAt,RedactedBy,RedactionKind,RedactionReason,State,TaskID'
      WHEN 'CALENDAR' THEN 'Audience,BoardID,CreatedAt,DeletedAt,Description,EndDate,EndsAt,OrganizationID,OwnerID,StartDate,StartsAt,Timezone,Title,UpdatedAt'
      WHEN 'RECURRENCE' THEN 'DTStart,EventID,HorizonEnd,RRULE,TZDBVersion,Timezone'
      WHEN 'FORM_LINK' THEN 'FormID,PermissionRevision,ReferenceEpoch,ResourceID,ResourceKind,VersionID'
      WHEN 'WORKFLOW_PROJECTION' THEN 'DueAt,PermissionRevision,Revision,SourceID,Summary,VersionID,VisibleToPrincipals'
      ELSE NULL END;
    IF v_keys IS DISTINCT FROM v_expected THEN RETURN false; END IF;
    IF p_target_kind = 'TASK' THEN
        RETURN pg_catalog.jsonb_typeof(p_record->'IsRecurring') = 'boolean'
           AND pg_catalog.jsonb_typeof(p_record->'RecurrenceRule') = 'string'
           AND pg_catalog.jsonb_typeof(p_record->'RecurrenceEnd') IN ('string','null')
           AND pg_catalog.char_length(p_record->>'RecurrenceRule') <= 50
           AND pg_catalog.octet_length((p_record->'FormFilter')::text) <= 128000;
    END IF;
    IF p_target_kind IS DISTINCT FROM 'WORKFLOW_PROJECTION' THEN RETURN true; END IF;
    IF pg_catalog.jsonb_typeof(p_record->'VisibleToPrincipals') IS DISTINCT FROM 'array' THEN RETURN false; END IF;
    RETURN work_projection_principals_valid_v1(ARRAY(SELECT value
      FROM pg_catalog.jsonb_array_elements_text(p_record->'VisibleToPrincipals') AS principal(value)
      ORDER BY value),true);
EXCEPTION WHEN invalid_parameter_value OR data_exception THEN RETURN false;
END;
$function$;
REVOKE ALL ON FUNCTION work_v1_adoption_document_valid_v1(text,jsonb) FROM PUBLIC;

CREATE FUNCTION work_v1_adoption_task_values_v2()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER
SET search_path FROM CURRENT AS $function$
DECLARE v_record jsonb; v_matches bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RETURN NEW;
    END IF;
    SELECT pg_catalog.jsonb_agg(stage.target_record ORDER BY stage.run_id)->0,
           pg_catalog.count(*) INTO v_record,v_matches
      FROM work_v1_adoption_stage AS stage
      JOIN work_adoption_runs AS run ON run.run_id=stage.run_id
     WHERE stage.target_kind='TASK' AND stage.target_id=NEW.id AND NOT stage.removed
       AND run.state IN ('COPYING','RECONCILING');
    IF v_matches IS DISTINCT FROM 1::bigint OR NOT work_v1_adoption_document_valid_v1('TASK',v_record) THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='Work V1 task compatibility evidence denied';
    END IF;
    NEW.creator_user_id=(v_record->>'CreatorID')::uuid;
    NEW.priority=v_record->>'Priority';
    NEW.starts_at=(v_record->>'StartsAt')::timestamptz;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION work_v1_adoption_task_values_v2() FROM PUBLIC;
CREATE TRIGGER work_v1_adoption_task_values_trigger_v2
    BEFORE INSERT OR UPDATE ON work_tasks
    FOR EACH ROW EXECUTE FUNCTION work_v1_adoption_task_values_v2();

CREATE FUNCTION work_v1_adoption_task_metadata_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER
SET search_path FROM CURRENT AS $function$
DECLARE v_record jsonb; v_matches bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RETURN NEW;
    END IF;
    SELECT pg_catalog.jsonb_agg(stage.target_record ORDER BY stage.run_id)->0,
           pg_catalog.count(*) INTO v_record,v_matches
      FROM work_v1_adoption_stage AS stage
      JOIN work_adoption_runs AS run ON run.run_id=stage.run_id
     WHERE stage.target_kind='TASK' AND stage.target_id=NEW.id AND NOT stage.removed
       AND run.state IN ('COPYING','RECONCILING');
    IF v_matches IS DISTINCT FROM 1::bigint OR NOT work_v1_adoption_document_valid_v1('TASK',v_record) THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='Work V1 task metadata evidence denied';
    END IF;
    INSERT INTO work_task_compatibility_metadata (
        task_id,is_recurring,recurrence_rule,recurrence_end,form_filter,created_at,updated_at
    ) VALUES (NEW.id,(v_record->>'IsRecurring')::boolean,v_record->>'RecurrenceRule',
        (v_record->>'RecurrenceEnd')::timestamptz,
        NULLIF(v_record->'FormFilter','null'::jsonb),NEW.created_at,NEW.updated_at)
    ON CONFLICT (task_id) DO UPDATE SET is_recurring=EXCLUDED.is_recurring,
        recurrence_rule=EXCLUDED.recurrence_rule,recurrence_end=EXCLUDED.recurrence_end,
        form_filter=EXCLUDED.form_filter,created_at=EXCLUDED.created_at,updated_at=EXCLUDED.updated_at;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION work_v1_adoption_task_metadata_v1() FROM PUBLIC;
CREATE TRIGGER work_v1_adoption_task_metadata_trigger_v1
    AFTER INSERT OR UPDATE ON work_tasks
    FOR EACH ROW EXECUTE FUNCTION work_v1_adoption_task_metadata_v1();
