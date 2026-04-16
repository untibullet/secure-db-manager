-- +goose Up
-- +goose StatementBegin

-------------------------------------------------------------------------------
-- 1. Таблица USERS
-------------------------------------------------------------------------------
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- 1.1. Чтение: Все видят активных коллег (нужно для UI)
CREATE POLICY users_select_active ON users
    FOR SELECT
    TO db_tester, db_developer, db_guest
    USING (is_active = TRUE);

-- 1.2. Полный доступ: Администратор (может создавать/удалять)
CREATE POLICY users_all_admin ON users
    FOR ALL
    TO db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

-- 1.3. Чтение всех: Лиды и Автоматизаторы (видят и уволенных)
CREATE POLICY users_select_all_staff ON users
    FOR SELECT
    TO db_test_lead, db_automation_engineer
    USING (TRUE);


-------------------------------------------------------------------------------
-- 2. Таблица TEST_CASES
-------------------------------------------------------------------------------
ALTER TABLE test_cases ENABLE ROW LEVEL SECURITY;

-- 2.1. Чтение: Обычные роли видят только активные кейсы
CREATE POLICY testcases_select_active ON test_cases
    FOR SELECT
    TO db_tester, db_developer, db_guest
    USING (is_active = TRUE);

-- 2.2. Полный доступ: Лиды, Автоматизаторы (могут править тесты)
CREATE POLICY testcases_all_staff ON test_cases
    FOR ALL
    TO db_test_lead, db_automation_engineer, db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

-- 2.3. Обновление: Тестировщик меняет ТОЛЬКО свои назначенные кейсы
CREATE POLICY testcases_update_own ON test_cases
    FOR UPDATE
    TO db_tester
    USING (
        -- Разрешено менять, если я - АВТОР (owner)
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_cases.owner_user_id
            AND u.username = CURRENT_USER
        )
    );


-------------------------------------------------------------------------------
-- 3. Таблица TEST_RESULTS
-------------------------------------------------------------------------------
ALTER TABLE test_results ENABLE ROW LEVEL SECURITY;

-- 3.1. Полный доступ: Лиды, Автоматизаторы, Админы
CREATE POLICY testresults_all_staff ON test_results
    FOR ALL
    TO db_test_lead, db_automation_engineer, db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

-- 3.2. Чтение СВОИХ результатов: Тестировщик (только записи, где executor = текущий пользователь)
CREATE POLICY testresults_select_tester ON test_results
    FOR SELECT
    TO db_tester
    USING (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_results.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

-- 3.3.1 Чтение всех результатов: Разработчик
CREATE POLICY testresults_select_developer ON test_results
    FOR SELECT
    TO db_developer
    USING (TRUE);

-- 3.3. Создание результатов: Тестировщик может INSERT только от своего имени
CREATE POLICY testresults_insert_own ON test_results
    FOR INSERT
    TO db_tester
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_results.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

-- 3.4. Обновление СВОИХ результатов: Тестировщик
CREATE POLICY testresults_update_own ON test_results
    FOR UPDATE
    TO db_tester
    USING (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_results.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

-- 3.5. Публичный доступ: Гости только видят завершенные тесты
CREATE POLICY testresults_select_public ON test_results
    FOR SELECT
    TO db_guest
    USING (
        EXISTS (
            SELECT 1 FROM statuses s 
            WHERE s.status_id = test_results.status_id 
            AND s.is_final = TRUE -- Просто "Завершенные"
        )
    );


-------------------------------------------------------------------------------
-- 4. Таблица AUDIT_LOG (ПолИБ-6.3)
-- Прямой доступ для всех ролей запрещён — никаких политик не добавляется.
-- INSERT выполняется только из SECURITY DEFINER функции audit_generic().
-- Чтение — исключительно через представление v_admin_audit_log для db_admin.
-------------------------------------------------------------------------------
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

CREATE OR REPLACE VIEW v_admin_audit_log AS
SELECT
    audit_id,
    table_name,
    operation,
    record_id,
    user_id,
    changed_at,
    old_values,
    new_values,
    ip_address
FROM audit_log
ORDER BY changed_at DESC;

GRANT SELECT ON v_admin_audit_log TO db_admin;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_admin_audit_log;
ALTER TABLE audit_log DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS users_select_active ON users;
DROP POLICY IF EXISTS users_all_admin ON users;
DROP POLICY IF EXISTS users_select_all_staff ON users;

DROP POLICY IF EXISTS testcases_select_active ON test_cases;
DROP POLICY IF EXISTS testcases_all_staff ON test_cases;
DROP POLICY IF EXISTS testcases_update_own ON test_cases;

DROP POLICY IF EXISTS testresults_all_staff ON test_results;
DROP POLICY IF EXISTS testresults_select_tester ON test_results;
DROP POLICY IF EXISTS testresults_select_developer ON test_results;
DROP POLICY IF EXISTS testresults_insert_own ON test_results;
DROP POLICY IF EXISTS testresults_update_own ON test_results;
DROP POLICY IF EXISTS testresults_select_public ON test_results;

ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_cases DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_results DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd
