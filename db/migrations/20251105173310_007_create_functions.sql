-- +goose Up
-- +goose StatementBegin
-- Функция обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Функция аудита для критических таблиц
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    user_id_value INT;
BEGIN
    -- Попытка получить user_id из настроек сессии
    BEGIN
        user_id_value := current_setting('app.current_user_id', true)::INT;
    EXCEPTION WHEN OTHERS THEN
        user_id_value := NULL;
    END;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (table_name, operation, record_id, user_id, new_values)
        VALUES (TG_TABLE_NAME, TG_OP, 
                CASE TG_TABLE_NAME
                    WHEN 'users' THEN NEW.user_id
                    WHEN 'test_cases' THEN NEW.test_case_id
                    WHEN 'test_plans' THEN NEW.test_plan_id
                    ELSE 0
                END,
                user_id_value, row_to_json(NEW)::JSONB);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values, new_values)
        VALUES (TG_TABLE_NAME, TG_OP,
                CASE TG_TABLE_NAME
                    WHEN 'users' THEN NEW.user_id
                    WHEN 'test_cases' THEN NEW.test_case_id
                    WHEN 'test_plans' THEN NEW.test_plan_id
                    ELSE 0
                END,
                user_id_value, row_to_json(OLD)::JSONB, row_to_json(NEW)::JSONB);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values)
        VALUES (TG_TABLE_NAME, TG_OP,
                CASE TG_TABLE_NAME
                    WHEN 'users' THEN OLD.user_id
                    WHEN 'test_cases' THEN OLD.test_case_id
                    WHEN 'test_plans' THEN OLD.test_plan_id
                    ELSE 0
                END,
                user_id_value, row_to_json(OLD)::JSONB);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Триггер для блокировки учетной записи после 5 неудачных попыток
CREATE OR REPLACE FUNCTION lock_user_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.failed_login_attempts >= 5 THEN
        NEW.account_locked_until = CURRENT_TIMESTAMP + INTERVAL '30 minutes';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_lock_user_account
    BEFORE UPDATE OF failed_login_attempts ON users
    FOR EACH ROW EXECUTE FUNCTION lock_user_account();

-- Функция для проверки сложности пароля
CREATE OR REPLACE FUNCTION check_password_complexity(password TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF LENGTH(password) < 12 THEN RETURN FALSE; END IF;
    IF password !~ '[A-Z]' THEN RETURN FALSE; END IF;
    IF password !~ '[a-z]' THEN RETURN FALSE; END IF;
    IF password !~ '[0-9]' THEN RETURN FALSE; END IF;
    IF password !~ '[!@#$%^&*()_+\-=\[\]{};:''",.<>?/|\\]' THEN RETURN FALSE; END IF;
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Функция для проверки истечения срока пароля
CREATE OR REPLACE FUNCTION check_password_expiration()
RETURNS TABLE(user_id INT, username VARCHAR, days_until_expiration INT) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.user_id,
        u.username,
        90 - EXTRACT(DAY FROM CURRENT_TIMESTAMP - u.password_changed_at)::INT AS days_until_expiration
    FROM users u
    WHERE u.is_active = TRUE
    AND CURRENT_TIMESTAMP - u.password_changed_at > INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Функция для просмотра активных сессий пользователей
CREATE OR REPLACE FUNCTION get_active_sessions()
RETURNS TABLE(
    pid INT,
    username TEXT,
    database TEXT,
    client_addr TEXT,
    application_name TEXT,
    state TEXT,
    query_start TIMESTAMPTZ,
    state_change TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pg_stat_activity.pid,
        pg_stat_activity.usename::TEXT,
        pg_stat_activity.datname::TEXT,
        COALESCE(pg_stat_activity.client_addr::TEXT, 'localhost'),
        pg_stat_activity.application_name::TEXT,
        pg_stat_activity.state::TEXT,
        pg_stat_activity.query_start,
        pg_stat_activity.state_change
    FROM pg_stat_activity
    WHERE pg_stat_activity.datname = current_database()
    AND pg_stat_activity.pid <> pg_backend_pid()
    ORDER BY pg_stat_activity.query_start DESC;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION get_active_sessions() IS 'Просмотр активных подключений к БД';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP FUNCTION IF EXISTS audit_trigger_function();

DROP FUNCTION IF EXISTS lock_user_account();

DROP FUNCTION IF EXISTS check_password_complexity(TEXT);

DROP FUNCTION IF EXISTS check_password_expiration(INT);

DROP FUNCTION IF EXISTS get_active_sessions();
-- +goose StatementEnd
