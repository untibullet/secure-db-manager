-- +goose Up
-- +goose StatementBegin
-- Включение RLS для пользователей
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_select_policy ON users
    FOR SELECT
    TO db_tester, db_developer, db_guest
    USING (is_active = TRUE);

CREATE POLICY users_select_all_policy ON users
    FOR SELECT
    TO db_test_lead, db_admin, db_automation_engineer
    USING (TRUE);

-- Включение RLS для тест-кейсов
ALTER TABLE test_cases ENABLE ROW LEVEL SECURITY;

CREATE POLICY testcases_select_policy ON test_cases
    FOR SELECT
    TO db_tester, db_developer, db_guest
    USING (is_active = TRUE);

CREATE POLICY testcases_select_all_policy ON test_cases
    FOR SELECT
    TO db_test_lead, db_admin, db_automation_engineer
    USING (TRUE);

-- Политика для тестировщика: видеть только назначенные тест-кейсы при UPDATE
CREATE POLICY testcases_update_assigned ON test_cases
    FOR UPDATE
    TO db_tester
    USING (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_cases.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

-- Включение RLS для результатов тестов
ALTER TABLE test_results ENABLE ROW LEVEL SECURITY;

CREATE POLICY testresults_select_own ON test_results
    FOR SELECT
    TO db_tester
    USING (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_results.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

CREATE POLICY testresults_select_all ON test_results
    FOR SELECT
    TO db_test_lead, db_admin, db_automation_engineer, db_developer
    USING (TRUE);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
-- Сначала удаляем политики
DROP POLICY IF EXISTS users_select_policy ON users;
DROP POLICY IF EXISTS users_select_all_policy ON users;
DROP POLICY IF EXISTS testcases_select_policy ON test_cases;
DROP POLICY IF EXISTS testcases_select_all_policy ON test_cases;
DROP POLICY IF EXISTS testcases_update_assigned ON test_cases;
DROP POLICY IF EXISTS testresults_select_own ON test_results;
DROP POLICY IF EXISTS testresults_select_all ON test_results;

-- Затем отключаем RLS для таблиц
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_cases DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_results DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd
