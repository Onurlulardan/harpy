-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_guard_board_column_limit_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE total bigint;
BEGIN
    PERFORM 1 FROM work_boards WHERE id = NEW.board_id FOR UPDATE;
    SELECT count(*) INTO total FROM work_board_columns WHERE board_id = NEW.board_id;
    IF total >= 100 THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'WORK_BOARD_LIMIT_EXCEEDED';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER work_board_columns_limit_guard
BEFORE INSERT ON work_board_columns
FOR EACH ROW EXECUTE FUNCTION work_guard_board_column_limit_v1();
ALTER TABLE work_board_columns ENABLE ALWAYS TRIGGER work_board_columns_limit_guard;
