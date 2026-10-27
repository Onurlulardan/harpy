-- +goose Up
CREATE TABLE iam_lifecycle_event_sequences (
    aggregate_kind text COLLATE "C" NOT NULL,
    aggregate_id uuid NOT NULL,
    next_sequence bigint NOT NULL,
    CONSTRAINT iam_lifecycle_event_sequences_pkey
        PRIMARY KEY (aggregate_kind, aggregate_id),
    CONSTRAINT iam_lifecycle_event_sequences_kind_check
        CHECK (aggregate_kind ~ '^[a-z][a-z0-9_.]{1,63}$'),
    CONSTRAINT iam_lifecycle_event_sequences_value_check CHECK (next_sequence > 0)
);

CREATE TABLE iam_lifecycle_event_receipts (
    operation_id uuid NOT NULL,
    aggregate_kind text COLLATE "C" NOT NULL,
    aggregate_id uuid NOT NULL,
    event_sequence bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_lifecycle_event_receipts_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_lifecycle_event_receipts_audit_fk FOREIGN KEY (operation_id)
        REFERENCES iam_lifecycle_audit (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_lifecycle_event_receipts_sequence_key
        UNIQUE (aggregate_kind, aggregate_id, event_sequence),
    CONSTRAINT iam_lifecycle_event_receipts_kind_check
        CHECK (aggregate_kind ~ '^[a-z][a-z0-9_.]{1,63}$'),
    CONSTRAINT iam_lifecycle_event_receipts_sequence_check
        CHECK (event_sequence > 0)
);

-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_event_sequence_v1(
    p_operation_id uuid, p_command_key text,
    p_aggregate_kind text, p_aggregate_id uuid
)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_sequence bigint;
    v_kind text;
    v_id uuid;
BEGIN
    SELECT receipt.aggregate_kind, receipt.aggregate_id, receipt.event_sequence
    INTO v_kind, v_id, v_sequence
    FROM iam_lifecycle_event_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_kind IS DISTINCT FROM p_aggregate_kind OR
           v_id IS DISTINCT FROM p_aggregate_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505';
        END IF;
        RETURN v_sequence;
    END IF;
    IF p_operation_id IS NULL OR p_aggregate_id IS NULL OR
       p_aggregate_kind !~ '^[a-z][a-z0-9_.]{1,63}$' OR
       NOT EXISTS (
          SELECT 1 FROM iam_lifecycle_audit AS audit
          WHERE audit.operation_id = p_operation_id
            AND audit.command_key = p_command_key
            AND audit.resource_id = p_aggregate_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501';
    END IF;
    INSERT INTO iam_lifecycle_event_sequences (
        aggregate_kind, aggregate_id, next_sequence
    ) VALUES (
        p_aggregate_kind, p_aggregate_id, 2
    ) ON CONFLICT (aggregate_kind, aggregate_id) DO UPDATE
      SET next_sequence = iam_lifecycle_event_sequences.next_sequence + 1
    RETURNING next_sequence - 1 INTO v_sequence;
    INSERT INTO iam_lifecycle_event_receipts (
        operation_id, aggregate_kind, aggregate_id, event_sequence, created_at
    ) VALUES (
        p_operation_id, p_aggregate_kind, p_aggregate_id,
        v_sequence, pg_catalog.transaction_timestamp()
    );
    RETURN v_sequence;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_event_sequence_v1(
    uuid,text,text,uuid
) FROM PUBLIC;

