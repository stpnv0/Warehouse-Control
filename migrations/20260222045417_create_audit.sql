-- +goose Up

-- ============================================================
-- Audit log
-- ============================================================
CREATE TABLE item_audit_log (
                                id         BIGSERIAL    PRIMARY KEY,
                                item_id    UUID         NOT NULL,
                                action     VARCHAR(10)  NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
                                changed_by UUID         NOT NULL, -- без FK
                                old_data   JSONB,
                                new_data   JSONB,
                                diff       JSONB,
                                changed_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_item_changed ON item_audit_log (item_id, changed_at DESC);
CREATE INDEX idx_audit_changed_at ON item_audit_log (changed_at DESC);
CREATE INDEX idx_audit_changed_by ON item_audit_log (changed_by);

-- Audit trigger function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_item_audit() RETURNS TRIGGER AS $$
DECLARE
    v_user_text TEXT;
    v_user UUID;
    v_old  JSONB;
    v_new  JSONB;
    v_diff JSONB;
    k      TEXT;
BEGIN
    -- Получаем ID пользователя
    v_user_text := current_setting('app.current_user_id', true);

    IF v_user_text IS NULL OR v_user_text = '' THEN
        RAISE EXCEPTION 'Audit Trigger Error: Session variable app.current_user_id is not set';
    END IF;

    BEGIN
        v_user := v_user_text::UUID;
    EXCEPTION WHEN invalid_text_representation THEN
        RAISE EXCEPTION 'Audit Trigger Error: Invalid UUID format in app.current_user_id: %', v_user_text;
    END;

    IF TG_OP = 'INSERT' THEN
        v_new := to_jsonb(NEW);
        INSERT INTO item_audit_log (item_id, action, changed_by, new_data)
        VALUES (NEW.id, 'INSERT', v_user, v_new);
        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        v_old  := to_jsonb(OLD);
        v_new  := to_jsonb(NEW);
        v_diff := '{}'::JSONB;

        FOR k IN SELECT jsonb_object_keys(v_new)
            LOOP
                -- Пропускаем служебные поля
                IF k IN ('id','updated_at', 'created_at') THEN
                    CONTINUE;
                END IF;

                IF (v_old -> k) IS DISTINCT FROM (v_new -> k) THEN
                    v_diff := v_diff || jsonb_build_object(
                            k, jsonb_build_object('old', v_old -> k, 'new', v_new -> k)
                                        );
                END IF;
            END LOOP;

        -- Если ничего не изменилось — не пишем
        IF v_diff = '{}'::JSONB THEN RETURN NEW;
        END IF;

        INSERT INTO item_audit_log (item_id, action, changed_by, old_data, new_data, diff)
        VALUES (NEW.id, 'UPDATE', v_user, v_old, v_new, v_diff);

        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        v_old := to_jsonb(OLD);
        INSERT INTO item_audit_log (item_id, action, changed_by, old_data)
        VALUES (OLD.id, 'DELETE', v_user, v_old);
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_item_audit
    AFTER INSERT OR UPDATE OR DELETE ON items
    FOR EACH ROW
EXECUTE FUNCTION fn_item_audit();

-- +goose Down
DROP TRIGGER IF EXISTS trg_item_audit ON items;
DROP FUNCTION IF EXISTS fn_item_audit();
DROP TABLE IF EXISTS item_audit_log;