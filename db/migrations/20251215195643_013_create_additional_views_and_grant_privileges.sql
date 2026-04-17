-- +goose Up
-- +goose StatementBegin

-- 1. Представление шагов тест-кейсов
CREATE OR REPLACE VIEW v_test_case_steps AS
SELECT
    tcs.step_id,
    tcs.test_case_id,
    tcs.step_order,
    tcs.action_text,
    tcs.expected_result,
    tcs.created_at
FROM test_case_steps tcs;

-- Права:
-- Test Lead, Automation Engineer, Developer: полный доступ на чтение
GRANT SELECT ON v_test_case_steps TO db_test_lead, db_automation_engineer, db_developer;
-- Tester: чтение (обычно они видят шаги при выполнении)
GRANT SELECT ON v_test_case_steps TO db_tester;
-- Test Lead, Automation Engineer: управление (для редактирования кейсов)
GRANT INSERT, UPDATE, DELETE ON v_test_case_steps TO db_test_lead, db_automation_engineer;
-- Tester: управление только своими кейсами (здесь упрощенно, обычно через триггеры/RLS на базовой таблице, но дадим права на view)
GRANT INSERT, UPDATE, DELETE ON v_test_case_steps TO db_tester; 


-- 2. Плоское представление autotests для записи (auto-updatable)
CREATE OR REPLACE VIEW v_autotests_manage AS
SELECT
    autotest_id,
    test_case_id,
    name,
    description,
    owner_user_id,
    is_active,
    created_at,
    updated_at
FROM autotests;

-- Права:
GRANT SELECT, INSERT, UPDATE, DELETE ON v_autotests_manage TO db_automation_engineer;
GRANT SELECT, UPDATE ON v_autotests_manage TO db_test_lead;


-- 3. Представление версий автотестов
CREATE OR REPLACE VIEW v_autotest_versions AS
SELECT 
    av.version_id,
    av.autotest_id,
    av.version_string,
    av.commit_hash,
    av.change_description,
    av.created_at
FROM autotest_versions av
JOIN autotests a ON av.autotest_id = a.autotest_id;

-- Права:
GRANT SELECT ON v_autotest_versions TO db_test_lead, db_automation_engineer, db_developer;
GRANT INSERT, UPDATE, DELETE ON v_autotest_versions TO db_automation_engineer; -- Основной владелец
GRANT INSERT, UPDATE, DELETE ON v_autotest_versions TO db_test_lead;


-- 3. Представление артефактов результатов
CREATE OR REPLACE VIEW v_test_result_artifacts AS
SELECT
    artifact_id,
    test_result_id,
    kind,
    file_path,
    file_size_bytes,
    mime_type,
    created_at
FROM test_result_artifacts;

-- Права:
GRANT SELECT ON v_test_result_artifacts TO db_test_lead, db_automation_engineer, db_developer, db_tester, db_guest;
-- Артефакты добавляют те, кто исполняет тесты (Tester, Automation Engineer, CI/CD)
GRANT INSERT, DELETE ON v_test_result_artifacts TO db_tester, db_automation_engineer, db_cicd_system;


-- 4. Представление элементов тестового прогона (связка Run <-> Case)
CREATE OR REPLACE VIEW v_test_run_items AS
SELECT
    run_item_id,
    test_run_id,
    test_case_id,
    execution_order,
    created_at
FROM test_run_items;

-- Права:
GRANT SELECT ON v_test_run_items TO db_test_lead, db_automation_engineer, db_developer, db_tester, db_guest;
-- Управляют составом рана обычно Лиды или CI/CD
GRANT INSERT, UPDATE, DELETE ON v_test_run_items TO db_test_lead, db_cicd_system;
-- Иногда тестировщики могут добавлять кейсы в ран
GRANT INSERT ON v_test_run_items TO db_tester;


-- 5. Плоское представление test_runs для записи (auto-updatable)
CREATE OR REPLACE VIEW v_test_runs_manage AS
SELECT
    test_run_id,
    test_plan_id,
    env_config_id,
    tool_config_id,
    version_id,
    name,
    description,
    start_date,
    end_date,
    status,
    created_at,
    created_by
FROM test_runs;

-- Права:
GRANT SELECT, INSERT, UPDATE, DELETE ON v_test_runs_manage TO db_test_lead;
GRANT SELECT, INSERT, UPDATE ON v_test_runs_manage TO db_cicd_system;


-- 6. Плоское представление users для записи (auto-updatable, без password_hash)
CREATE OR REPLACE VIEW v_users_manage AS
SELECT
    user_id,
    username,
    email,
    full_name,
    is_active,
    failed_login_attempts,
    account_locked_until,
    password_changed_at,
    last_login,
    created_at,
    updated_at
FROM users;

-- Права: только администратор управляет пользователями
GRANT SELECT, INSERT, UPDATE, DELETE ON v_users_manage TO db_admin;


-- 7. Представление ролей пользователей (для Админа)
-- Примечание: v_admin_users_and_roles уже существует для просмотра, это view для M2M операций
CREATE OR REPLACE VIEW v_user_roles_manage AS
SELECT
    ur.user_id,
    ur.role_id,
    r.name AS role_name,
    ur.granted_at,
    ur.valid_until
FROM user_roles ur
JOIN roles r ON ur.role_id = r.role_id;

-- Права:
GRANT SELECT, INSERT, UPDATE, DELETE ON v_user_roles_manage TO db_admin;


-- 8. Плоское представление test_results для записи (auto-updatable, AD-4)
CREATE OR REPLACE VIEW v_results_manage AS
SELECT
    test_result_id,
    run_item_id,
    status_id,
    executor_user_id,
    execution_date,
    result_summary,
    execution_duration_minutes,
    error_message,
    created_at
FROM test_results;

-- Права:
GRANT SELECT, INSERT, UPDATE, DELETE ON v_results_manage TO db_test_lead;
GRANT SELECT, INSERT, UPDATE ON v_results_manage TO db_tester, db_automation_engineer, db_cicd_system;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_test_case_steps;
DROP VIEW IF EXISTS v_autotests_manage;
DROP VIEW IF EXISTS v_autotest_versions;
DROP VIEW IF EXISTS v_test_result_artifacts;
DROP VIEW IF EXISTS v_test_run_items;
DROP VIEW IF EXISTS v_test_runs_manage;
DROP VIEW IF EXISTS v_users_manage;
DROP VIEW IF EXISTS v_user_roles_manage;
DROP VIEW IF EXISTS v_results_manage;
-- +goose StatementEnd
