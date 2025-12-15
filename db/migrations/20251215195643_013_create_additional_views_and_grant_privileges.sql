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
FROM test_case_steps tcs
JOIN test_cases tc ON tcs.test_case_id = tc.test_case_id;

-- Права:
-- Test Lead, Automation Engineer, Developer: полный доступ на чтение
GRANT SELECT ON v_test_case_steps TO db_test_lead, db_automation_engineer, db_developer;
-- Tester: чтение (обычно они видят шаги при выполнении)
GRANT SELECT ON v_test_case_steps TO db_tester;
-- Test Lead, Automation Engineer: управление (для редактирования кейсов)
GRANT INSERT, UPDATE, DELETE ON v_test_case_steps TO db_test_lead, db_automation_engineer;
-- Tester: управление только своими кейсами (здесь упрощенно, обычно через триггеры/RLS на базовой таблице, но дадим права на view)
GRANT INSERT, UPDATE, DELETE ON v_test_case_steps TO db_tester; 


-- 2. Представление версий автотестов
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
    tra.artifact_id,
    tra.test_result_id,
    tra.kind,
    tra.file_path,
    tra.file_size_bytes,
    tra.mime_type,
    tra.created_at
FROM test_result_artifacts tra
JOIN test_results tr ON tra.test_result_id = tr.test_result_id;

-- Права:
GRANT SELECT ON v_test_result_artifacts TO db_test_lead, db_automation_engineer, db_developer, db_tester, db_guest;
-- Артефакты добавляют те, кто исполняет тесты (Tester, Automation Engineer, CI/CD)
GRANT INSERT, DELETE ON v_test_result_artifacts TO db_tester, db_automation_engineer, db_cicd_system;


-- 4. Представление элементов тестового прогона (связка Run <-> Case)
CREATE OR REPLACE VIEW v_test_run_items AS
SELECT 
    tri.run_item_id,
    tri.test_run_id,
    tri.test_case_id,
    tri.execution_order,
    tri.created_at
FROM test_run_items tri
JOIN test_runs tr ON tri.test_run_id = tr.test_run_id;

-- Права:
GRANT SELECT ON v_test_run_items TO db_test_lead, db_automation_engineer, db_developer, db_tester, db_guest;
-- Управляют составом рана обычно Лиды или CI/CD
GRANT INSERT, UPDATE, DELETE ON v_test_run_items TO db_test_lead, db_cicd_system;
-- Иногда тестировщики могут добавлять кейсы в ран
GRANT INSERT ON v_test_run_items TO db_tester;


-- 5. Представление ролей пользователей (для Админа)
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_test_case_steps;
DROP VIEW IF EXISTS v_autotest_versions;
DROP VIEW IF EXISTS v_test_result_artifacts;
DROP VIEW IF EXISTS v_test_run_items;
DROP VIEW IF EXISTS v_user_roles_manage;
-- +goose StatementEnd
