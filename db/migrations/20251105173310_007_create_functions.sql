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

-- Универсальная функция аудита (ПолИБ-6.2).
-- Используется всеми таблицами активов A1–A7.
-- TG_ARGV[0] — имя колонки первичного ключа таблицы (передаётся при создании триггера).
CREATE OR REPLACE FUNCTION audit_generic()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id   INT;
    v_record_id INT;
BEGIN
    BEGIN
        v_user_id := current_setting('app.current_user_id', true)::INT;
    EXCEPTION WHEN OTHERS THEN
        v_user_id := NULL;
    END;

    IF TG_OP = 'DELETE' THEN
        EXECUTE format('SELECT ($1).%I', TG_ARGV[0]) INTO v_record_id USING OLD;
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values, new_values, ip_address)
        VALUES (TG_TABLE_NAME, TG_OP, v_record_id, v_user_id, to_jsonb(OLD), NULL, inet_client_addr());
        RETURN OLD;
    ELSIF TG_OP = 'INSERT' THEN
        EXECUTE format('SELECT ($1).%I', TG_ARGV[0]) INTO v_record_id USING NEW;
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values, new_values, ip_address)
        VALUES (TG_TABLE_NAME, TG_OP, v_record_id, v_user_id, NULL, to_jsonb(NEW), inet_client_addr());
        RETURN NEW;
    ELSE -- UPDATE
        EXECUTE format('SELECT ($1).%I', TG_ARGV[0]) INTO v_record_id USING NEW;
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values, new_values, ip_address)
        VALUES (TG_TABLE_NAME, TG_OP, v_record_id, v_user_id, to_jsonb(OLD), to_jsonb(NEW), inet_client_addr());
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public;

COMMENT ON FUNCTION audit_generic() IS 'Универсальный аудит-триггер. SECURITY DEFINER позволяет INSERT в audit_log в обход RLS.';

-- Триггер для блокировки учетной записи по порогам неудачных попыток (ПолИБ-2.5):
--   >= 5  попыток → блокировка на 15 минут
--   >= 10 попыток → бессрочная блокировка до ручной разблокировки администратором
CREATE OR REPLACE FUNCTION lock_user_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.failed_login_attempts > OLD.failed_login_attempts THEN
        IF NEW.failed_login_attempts >= 10 THEN
            NEW.account_locked_until = 'infinity'::TIMESTAMPTZ;
        ELSIF NEW.failed_login_attempts >= 5 THEN
            NEW.account_locked_until = CURRENT_TIMESTAMP + INTERVAL '15 minutes';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_lock_user_account
BEFORE UPDATE OF failed_login_attempts ON users
FOR EACH ROW
WHEN (OLD.failed_login_attempts IS DISTINCT FROM NEW.failed_login_attempts)
EXECUTE FUNCTION lock_user_account();

-- Функция для просмотра активных сессий пользователей
CREATE OR REPLACE FUNCTION get_active_sessions()
RETURNS TABLE(
    pid             INT,
    username        TEXT,
    database        TEXT,
    client_addr     TEXT,
    application_name TEXT,
    state           TEXT,
    query_start     TIMESTAMPTZ,
    state_change    TIMESTAMPTZ
) AS $$
BEGIN
    -- Защита от подмены search_path
    PERFORM set_config('search_path', 'pg_catalog,public', true);

    RETURN QUERY
    SELECT
        a.pid,
        a.usename::TEXT,
        a.datname::TEXT,
        COALESCE(a.client_addr::TEXT, 'localhost'),
        a.application_name::TEXT,
        a.state::TEXT,
        a.query_start,
        a.state_change
    FROM pg_catalog.pg_stat_activity AS a
    WHERE a.datname = current_database()
      AND a.pid <> pg_backend_pid()
    ORDER BY a.query_start DESC;
END;
$$ LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public;

COMMENT ON FUNCTION get_active_sessions() IS 'Просмотр активных подключений к БД';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS audit_generic();
DROP TRIGGER IF EXISTS trg_lock_user_account ON users;
DROP FUNCTION IF EXISTS lock_user_account();
DROP FUNCTION IF EXISTS get_active_sessions();
-- +goose StatementEnd
