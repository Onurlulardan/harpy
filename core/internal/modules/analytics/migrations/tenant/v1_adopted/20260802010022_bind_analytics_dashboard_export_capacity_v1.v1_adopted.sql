-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_capacity_transition_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_admission bigint;
    v_tenant_running bigint;
    v_principal_admission bigint;
    v_reservation analytics_dashboard_export_capacity_reservations%ROWTYPE;
BEGIN
    IF TG_OP='INSERT' THEN
        IF NEW.state<>'queued' THEN
            RAISE EXCEPTION USING ERRCODE='P6112',
                MESSAGE='Analytics dashboard export admission state is invalid';
        END IF;
        INSERT INTO analytics_dashboard_export_capacity_counters(
            tenant_id,updated_at
        ) VALUES (NEW.tenant_id,NEW.updated_at) ON CONFLICT DO NOTHING;
        SELECT admission_count,running_count
        INTO v_tenant_admission,v_tenant_running
        FROM analytics_dashboard_export_capacity_counters
        WHERE tenant_id=NEW.tenant_id FOR UPDATE;
        IF v_tenant_admission>=20 THEN
            RAISE EXCEPTION USING ERRCODE='P6110',
                MESSAGE='Analytics dashboard export tenant admission is full';
        END IF;
        INSERT INTO analytics_dashboard_export_principal_counters(
            actor_id,updated_at
        ) VALUES (NEW.actor_id,NEW.updated_at) ON CONFLICT DO NOTHING;
        SELECT admission_count INTO v_principal_admission
        FROM analytics_dashboard_export_principal_counters
        WHERE actor_id=NEW.actor_id FOR UPDATE;
        IF v_principal_admission>=5 THEN
            RAISE EXCEPTION USING ERRCODE='P6110',
                MESSAGE='Analytics dashboard export principal admission is full';
        END IF;
        UPDATE analytics_dashboard_export_capacity_counters
        SET admission_count=admission_count+1,revision=revision+1,
            updated_at=GREATEST(updated_at+interval '1 microsecond',NEW.updated_at)
        WHERE tenant_id=NEW.tenant_id;
        UPDATE analytics_dashboard_export_principal_counters
        SET admission_count=admission_count+1,revision=revision+1,
            updated_at=GREATEST(updated_at+interval '1 microsecond',NEW.updated_at)
        WHERE actor_id=NEW.actor_id;
        INSERT INTO analytics_dashboard_export_capacity_reservations(
            export_id,tenant_id,actor_id,admission_held,running_held,
            acquired_at,updated_at
        ) VALUES (
            NEW.export_id,NEW.tenant_id,NEW.actor_id,true,false,
            NEW.created_at,NEW.updated_at
        );
        INSERT INTO analytics_dashboard_export_capacity_receipts(
            export_id,transition_kind,attempt_generation,tenant_id,actor_id,occurred_at
        ) VALUES (
            NEW.export_id,'admission_acquired',0,NEW.tenant_id,
            NEW.actor_id,NEW.updated_at
        );
        RETURN NEW;
    END IF;

    SELECT reservation.* INTO v_reservation
    FROM analytics_dashboard_export_capacity_reservations AS reservation
    WHERE reservation.export_id=NEW.export_id FOR UPDATE;
    IF v_reservation.export_id IS NULL THEN
        IF OLD.state NOT IN ('queued','running') AND
           NEW.state NOT IN ('queued','running') THEN
            RETURN NEW;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6112',
            MESSAGE='Analytics dashboard export capacity reservation is missing';
    END IF;

    IF OLD.state<>'running' AND NEW.state='running' THEN
        SELECT admission_count,running_count
        INTO v_tenant_admission,v_tenant_running
        FROM analytics_dashboard_export_capacity_counters
        WHERE tenant_id=NEW.tenant_id FOR UPDATE;
        IF NOT v_reservation.admission_held OR v_reservation.running_held OR
           v_tenant_running>=2 THEN
            RAISE EXCEPTION USING ERRCODE='P6111',
                MESSAGE='Analytics dashboard export running capacity is full';
        END IF;
        UPDATE analytics_dashboard_export_capacity_counters
        SET running_count=running_count+1,revision=revision+1,
            updated_at=GREATEST(updated_at+interval '1 microsecond',NEW.updated_at)
        WHERE tenant_id=NEW.tenant_id;
        UPDATE analytics_dashboard_export_capacity_reservations
        SET running_held=true,running_generation=NEW.attempt_generation,
            updated_at=NEW.updated_at WHERE export_id=NEW.export_id;
        INSERT INTO analytics_dashboard_export_capacity_receipts(
            export_id,transition_kind,attempt_generation,tenant_id,actor_id,occurred_at
        ) VALUES (
            NEW.export_id,'running_acquired',NEW.attempt_generation,
            NEW.tenant_id,NEW.actor_id,NEW.updated_at
        );
    END IF;

    IF OLD.state='running' AND NEW.state<>'running' THEN
        IF NOT v_reservation.running_held OR
           v_reservation.running_generation<>OLD.attempt_generation THEN
            RAISE EXCEPTION USING ERRCODE='P6112',
                MESSAGE='Analytics dashboard export running reservation differs';
        END IF;
        UPDATE analytics_dashboard_export_capacity_counters
        SET running_count=running_count-1,revision=revision+1,
            updated_at=GREATEST(updated_at+interval '1 microsecond',NEW.updated_at)
        WHERE tenant_id=NEW.tenant_id AND running_count>0;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6112',
                MESSAGE='Analytics dashboard export running counter underflow';
        END IF;
        UPDATE analytics_dashboard_export_capacity_reservations
        SET running_held=false,running_generation=NULL,updated_at=NEW.updated_at
        WHERE export_id=NEW.export_id;
        INSERT INTO analytics_dashboard_export_capacity_receipts(
            export_id,transition_kind,attempt_generation,tenant_id,actor_id,occurred_at
        ) VALUES (
            NEW.export_id,'running_released',OLD.attempt_generation,
            NEW.tenant_id,NEW.actor_id,NEW.updated_at
        );
    END IF;

    IF OLD.state IN ('queued','running') AND
       NEW.state NOT IN ('queued','running') THEN
        IF NOT v_reservation.admission_held THEN
            RAISE EXCEPTION USING ERRCODE='P6112',
                MESSAGE='Analytics dashboard export admission reservation differs';
        END IF;
        UPDATE analytics_dashboard_export_capacity_counters
        SET admission_count=admission_count-1,revision=revision+1,
            updated_at=GREATEST(updated_at+interval '1 microsecond',NEW.updated_at)
        WHERE tenant_id=NEW.tenant_id AND admission_count>0;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6112',
                MESSAGE='Analytics dashboard export tenant admission counter underflow';
        END IF;
        UPDATE analytics_dashboard_export_principal_counters
        SET admission_count=admission_count-1,revision=revision+1,
            updated_at=GREATEST(updated_at+interval '1 microsecond',NEW.updated_at)
        WHERE actor_id=NEW.actor_id AND admission_count>0;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6112',
                MESSAGE='Analytics dashboard export admission counter underflow';
        END IF;
        UPDATE analytics_dashboard_export_capacity_reservations
        SET admission_held=false,updated_at=NEW.updated_at
        WHERE export_id=NEW.export_id;
        INSERT INTO analytics_dashboard_export_capacity_receipts(
            export_id,transition_kind,attempt_generation,tenant_id,actor_id,occurred_at
        ) VALUES (
            NEW.export_id,'admission_released',0,NEW.tenant_id,
            NEW.actor_id,NEW.updated_at
        );
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER analytics_dashboard_exports_capacity_transition
AFTER INSERT OR UPDATE OF state,attempt_generation
ON analytics_dashboard_exports
FOR EACH ROW EXECUTE FUNCTION analytics_dashboard_export_capacity_transition_v1();
ALTER TABLE analytics_dashboard_exports
    ENABLE ALWAYS TRIGGER analytics_dashboard_exports_capacity_transition;
REVOKE ALL ON FUNCTION analytics_dashboard_export_capacity_transition_v1()
FROM PUBLIC;
