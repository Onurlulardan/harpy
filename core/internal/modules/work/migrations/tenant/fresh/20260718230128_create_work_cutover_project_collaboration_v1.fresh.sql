-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_project_collaboration_v1(
    p_target_kind text,
    p_target_id uuid,
    p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
SET timezone = 'UTC'
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       NOT work_v1_adoption_document_valid_v1(p_target_kind, p_record) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Work V1 collaboration projection denied';
    END IF;
    CASE p_target_kind
    WHEN 'COMMENT' THEN
        INSERT INTO work_comments (
            id, board_id, task_id, author_id, body, state, redaction_kind,
            redacted_by, redaction_reason, row_version, created_at, redacted_at
        ) VALUES (
            p_target_id, (p_record->>'BoardID')::uuid,
            (p_record->>'TaskID')::uuid, (p_record->>'AuthorID')::uuid,
            p_record->>'Body', p_record->>'State',
            NULLIF(p_record->>'RedactionKind',''),
            NULLIF(p_record->>'RedactedBy','')::uuid,
            NULLIF(p_record->>'RedactionReason',''), 1,
            (p_record->>'CreatedAt')::timestamptz,
            (p_record->>'RedactedAt')::timestamptz
        );
    WHEN 'CALENDAR' THEN
        INSERT INTO work_calendar_events (
            id, owner_id, audience, board_id, organization_id, title,
            description, starts_at, ends_at, start_date, end_date,
            timezone_name, generated, edited, live_reference_count, legal_hold,
            row_version, deleted_at, created_at, updated_at
        ) VALUES (
            p_target_id, (p_record->>'OwnerID')::uuid, p_record->>'Audience',
            NULLIF(p_record->>'BoardID','')::uuid,
            NULLIF(p_record->>'OrganizationID','')::uuid,
            p_record->>'Title', p_record->>'Description',
            (p_record->>'StartsAt')::timestamptz,
            (p_record->>'EndsAt')::timestamptz,
            NULLIF(p_record->>'StartDate','')::date,
            NULLIF(p_record->>'EndDate','')::date,
            p_record->>'Timezone', false, false, 0, false, 1,
            (p_record->>'DeletedAt')::timestamptz,
            (p_record->>'CreatedAt')::timestamptz,
            (p_record->>'UpdatedAt')::timestamptz
        )
        ON CONFLICT (id) DO UPDATE SET
            owner_id=EXCLUDED.owner_id, audience=EXCLUDED.audience,
            board_id=EXCLUDED.board_id,
            organization_id=EXCLUDED.organization_id, title=EXCLUDED.title,
            description=EXCLUDED.description, starts_at=EXCLUDED.starts_at,
            ends_at=EXCLUDED.ends_at, start_date=EXCLUDED.start_date,
            end_date=EXCLUDED.end_date, timezone_name=EXCLUDED.timezone_name,
            deleted_at=EXCLUDED.deleted_at, created_at=EXCLUDED.created_at,
            updated_at=EXCLUDED.updated_at;
    WHEN 'RECURRENCE' THEN
        INSERT INTO work_recurrence_rules (
            id, owner_event_id, current_revision, lifecycle_state, row_version,
            created_at, updated_at
        ) VALUES (
            p_target_id, (p_record->>'EventID')::uuid, 1, 'ACTIVE', 1,
            transaction_timestamp(), transaction_timestamp()
        )
        ON CONFLICT (id) DO UPDATE SET
            owner_event_id=EXCLUDED.owner_event_id, current_revision=1,
            lifecycle_state='ACTIVE', updated_at=EXCLUDED.updated_at;
        INSERT INTO work_recurrence_rule_revisions (
            rule_id, revision, contract_version, dtstart, timezone_name,
            tzdb_version, rrule, exdates, effective_from, created_at
        ) VALUES (
            p_target_id, 1, 'work_rrule_v1', p_record->>'DTStart',
            p_record->>'Timezone', p_record->>'TZDBVersion',
            p_record->>'RRULE', ARRAY[]::text[],
            pg_catalog.to_timestamp(
                p_record->>'DTStart', 'YYYYMMDD"T"HH24MISS'
            ), transaction_timestamp()
        )
        ON CONFLICT (rule_id, revision) DO NOTHING;
        IF NOT EXISTS (
            SELECT 1 FROM work_recurrence_rule_revisions AS revision
            WHERE revision.rule_id=p_target_id AND revision.revision=1
              AND revision.dtstart=p_record->>'DTStart'
              AND revision.timezone_name=p_record->>'Timezone'
              AND revision.tzdb_version=p_record->>'TZDBVersion'
              AND revision.rrule=p_record->>'RRULE'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'Work V1 recurrence projection differs';
        END IF;
    WHEN 'WORKFLOW_PROJECTION' THEN
        INSERT INTO work_inbox_items (
            id, source_owner, source_kind, source_id, source_version_id,
            source_revision, permission_revision, event_sequence, summary,
            due_at, state, updated_at
        ) VALUES (
            p_target_id, 'WORKFLOW', 'HUMAN_TASK',
            (p_record->>'SourceID')::uuid, (p_record->>'VersionID')::uuid,
            (p_record->>'Revision')::bigint,
            (p_record->>'PermissionRevision')::bigint,
            (p_record->>'Revision')::bigint, p_record->>'Summary',
            (p_record->>'DueAt')::timestamptz,
            'VISIBLE', transaction_timestamp()
        )
        ON CONFLICT (id) DO UPDATE SET
            source_version_id=EXCLUDED.source_version_id,
            source_revision=EXCLUDED.source_revision,
            permission_revision=EXCLUDED.permission_revision,
            event_sequence=EXCLUDED.event_sequence,
            summary=EXCLUDED.summary, due_at=EXCLUDED.due_at,
            state='VISIBLE', updated_at=EXCLUDED.updated_at;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Work V1 collaboration projection kind is invalid';
    END CASE;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_project_collaboration_v1(
    text, uuid, jsonb
) FROM PUBLIC;

