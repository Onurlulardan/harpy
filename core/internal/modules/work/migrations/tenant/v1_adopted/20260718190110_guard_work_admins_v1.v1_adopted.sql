-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_guard_last_user_admin_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE remaining bigint;
BEGIN
    IF OLD.active AND OLD.principal_kind = 'USER' AND OLD.board_role = 'ADMIN'
       AND (TG_OP = 'DELETE' OR NOT NEW.active OR NEW.principal_kind <> 'USER'
            OR NEW.board_role <> 'ADMIN') THEN
        PERFORM 1 FROM work_boards WHERE id = OLD.board_id FOR UPDATE;
        SELECT count(*) INTO remaining FROM work_board_memberships
         WHERE board_id = OLD.board_id AND active AND principal_kind = 'USER'
           AND board_role = 'ADMIN' AND id <> OLD.id;
        IF remaining = 0 THEN
            RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'WORK_LAST_USER_ADMIN';
        END IF;
    END IF;
    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER work_memberships_admin_guard
BEFORE UPDATE OR DELETE ON work_board_memberships
FOR EACH ROW EXECUTE FUNCTION work_guard_last_user_admin_v1();
ALTER TABLE work_board_memberships ENABLE ALWAYS TRIGGER work_memberships_admin_guard;
