-- +goose Up
-- +goose StatementBegin
-- 1. Представление публичных тест-планов
CREATE OR REPLACE VIEW v_public_test_plans AS
SELECT
    tp.test_plan_id,
    tp.name AS test_plan_name,
    tp.description,
    tp.start_date,
    tp.end_date,
    tp.status,
    u.full_name AS owner,
    p.name AS priority,
    p.rank AS priority_rank
FROM test_plans tp
JOIN users u ON tp.owner_user_id = u.user_id
JOIN priorities p ON tp.priority_id = p.priority_id
WHERE tp.status IN ('ACTIVE', 'COMPLETED');

COMMENT ON VIEW v_public_test_plans IS 'Публичные тест-планы (доступны всем ролям)';

-- 2. Представление назначенных тест-кейсов для тестировщика
CREATE OR REPLACE VIEW v_tester_assigned_cases AS
SELECT
    tc.test_case_id,
    tc.name AS test_case_name,
    tc.description,
    tc.is_automated,
    p.name AS priority,
    u_owner.full_name AS owner,
    u_executor.username AS executor_username
FROM test_cases tc
JOIN priorities p ON tc.priority_id = p.priority_id
JOIN users u_owner ON tc.owner_user_id = u_owner.user_id
JOIN users u_executor ON tc.executor_user_id = u_executor.user_id
WHERE u_executor.username = CURRENT_USER
AND tc.is_active = TRUE;

COMMENT ON VIEW v_tester_assigned_cases IS 'Тест-кейсы, назначенные текущему пользователю (RLS)';

-- 3. Представление всех тест-кейсов для тест-лида
CREATE OR REPLACE VIEW v_lead_all_cases AS
SELECT
    tc.test_case_id,
    tc.name AS test_case_name,
    tc.description,
    tc.is_automated,
    tc.is_active,
    p.name AS priority,
    u_owner.full_name AS owner,
    COALESCE(u_executor.full_name, 'Не назначен') AS executor
FROM test_cases tc
JOIN priorities p ON tc.priority_id = p.priority_id
JOIN users u_owner ON tc.owner_user_id = u_owner.user_id
LEFT JOIN users u_executor ON tc.executor_user_id = u_executor.user_id;

COMMENT ON VIEW v_lead_all_cases IS 'Все тест-кейсы (доступ: Test_Lead, Automation_Engineer)';

-- 4. Представление автотестов для инженера по автоматизации
CREATE OR REPLACE VIEW v_engineer_autotests AS
SELECT
    a.autotest_id,
    a.name AS autotest_name,
    a.description,
    a.is_active,
    av.version_id,
    av.version_string,
    av.commit_hash,
    av.created_at AS version_date,
    u.full_name AS author
FROM autotests a
JOIN autotest_versions av ON a.autotest_id = av.autotest_id
JOIN users u ON a.owner_user_id = u.user_id
WHERE a.is_active = TRUE;

COMMENT ON VIEW v_engineer_autotests IS 'Автотесты с версиями (доступ: Automation_Engineer, Test_Lead)';

-- 5. Представление конфигураций окружений
CREATE OR REPLACE VIEW v_shared_environments AS
SELECT
    ec.env_config_id,
    ec.name AS environment_name,
    ec.description,
    ec.is_active,
    COUNT(ecp.env_config_param_id) AS params_count
FROM env_configs ec
LEFT JOIN env_config_params ecp ON ec.env_config_id = ecp.env_config_id
WHERE ec.is_active = TRUE
GROUP BY ec.env_config_id, ec.name, ec.description, ec.is_active;

COMMENT ON VIEW v_shared_environments IS 'Конфигурации окружений (доступны многим ролям)';

-- 6. Представление результатов тестов текущего пользователя
CREATE OR REPLACE VIEW v_tester_my_results AS
SELECT
    tr.test_result_id,
    tri.test_run_id,
    tc.name AS test_case_name,
    s.name AS status,
    tr.execution_date,
    tr.result_summary,
    tr.execution_duration_minutes,
    u.username AS executor_username
FROM test_results tr
JOIN test_run_items tri ON tr.run_item_id = tri.run_item_id
JOIN test_cases tc ON tri.test_case_id = tc.test_case_id
JOIN statuses s ON tr.status_id = s.status_id
JOIN users u ON tr.executor_user_id = u.user_id
WHERE u.username = CURRENT_USER;

COMMENT ON VIEW v_tester_my_results IS 'Результаты тестов текущего пользователя (RLS)';

-- 7. Представление всех результатов для тест-лида
CREATE OR REPLACE VIEW v_lead_all_results AS
SELECT
    tr.test_result_id,
    tri.test_run_id,
    trun.name AS test_run_name,
    tc.name AS test_case_name,
    s.name AS status,
    COALESCE(u.full_name, 'Автоматический запуск') AS executor,
    tr.execution_date,
    tr.result_summary,
    tr.execution_duration_minutes,
    tr.error_message
FROM test_results tr
JOIN test_run_items tri ON tr.run_item_id = tri.run_item_id
JOIN test_runs trun ON tri.test_run_id = trun.test_run_id
JOIN test_cases tc ON tri.test_case_id = tc.test_case_id
JOIN statuses s ON tr.status_id = s.status_id
LEFT JOIN users u ON tr.executor_user_id = u.user_id;

COMMENT ON VIEW v_lead_all_results IS 'Все результаты тестов (доступ: Test_Lead, Automation_Engineer)';

-- 8. Представление публичных результатов для гостей
CREATE OR REPLACE VIEW v_public_results AS
SELECT
    tr.test_result_id,
    tri.test_run_id,
    tc.name AS test_case_name,
    s.name AS status,
    tr.execution_date
FROM test_results tr
JOIN test_run_items tri ON tr.run_item_id = tri.run_item_id
JOIN test_cases tc ON tri.test_case_id = tc.test_case_id
JOIN statuses s ON tr.status_id = s.status_id
WHERE s.code IN ('PASSED', 'FAILED', 'BLOCKED')
AND s.is_final = TRUE;

COMMENT ON VIEW v_public_results IS 'Публичные результаты тестов (доступ: Guest)';

-- 9. Представление отчетов о качестве
CREATE OR REPLACE VIEW v_shared_reports AS
SELECT
    r.report_id,
    r.name AS report_name,
    r.created_at AS creation_date,
    rt.name AS template,
    u.full_name AS author
FROM reports r
JOIN report_templates rt ON r.template_id = rt.template_id
JOIN users u ON r.owner_user_id = u.user_id
WHERE rt.is_active = TRUE;

COMMENT ON VIEW v_shared_reports IS 'Отчеты о качестве (доступны многим ролям)';

-- 10. Представление пользователей и ролей для администратора
CREATE OR REPLACE VIEW v_admin_users_and_roles AS
SELECT
    u.user_id,
    u.username,
    u.full_name,
    u.email,
    u.is_active,
    u.created_at,
    u.last_login,
    u.failed_login_attempts,
    STRING_AGG(r.name, ', ') AS roles
FROM users u
LEFT JOIN user_roles ur ON u.user_id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.role_id
GROUP BY u.user_id, u.username, u.full_name, u.email, u.is_active, 
         u.created_at, u.last_login, u.failed_login_attempts;

COMMENT ON VIEW v_admin_users_and_roles IS 'Управление пользователями и ролями (доступ: Administrator)';

-- 11. Представление активных тестовых прогонов с статистикой
CREATE OR REPLACE VIEW v_active_test_runs AS
SELECT 
    tr.test_run_id,
    tr.name AS run_name,
    tp.name AS plan_name,
    av.version_string,
    ec.name AS environment,
    tc.name AS tool_config,
    tr.status,
    tr.start_date,
    tr.end_date,
    u.full_name AS created_by_name,
    COUNT(DISTINCT tri.test_case_id) AS total_tests,
    COUNT(DISTINCT CASE WHEN ts.code = 'PASSED' THEN tres.test_result_id END) AS passed_tests,
    COUNT(DISTINCT CASE WHEN ts.code = 'FAILED' THEN tres.test_result_id END) AS failed_tests,
    COUNT(DISTINCT CASE WHEN ts.code = 'BLOCKED' THEN tres.test_result_id END) AS blocked_tests
FROM test_runs tr
JOIN test_plans tp ON tr.test_plan_id = tp.test_plan_id
JOIN app_versions av ON tr.version_id = av.version_id
JOIN env_configs ec ON tr.env_config_id = ec.env_config_id
JOIN tool_configs tc ON tr.tool_config_id = tc.tool_config_id
LEFT JOIN users u ON tr.created_by = u.user_id
LEFT JOIN test_run_items tri ON tr.test_run_id = tri.test_run_id
LEFT JOIN test_results tres ON tri.run_item_id = tres.run_item_id
LEFT JOIN statuses ts ON tres.status_id = ts.status_id
WHERE tr.status IN ('PLANNED', 'IN_PROGRESS')
GROUP BY tr.test_run_id, tr.name, tp.name, av.version_string, 
         ec.name, tc.name, tr.status, tr.start_date, tr.end_date, u.full_name;

COMMENT ON VIEW v_active_test_runs IS 'Активные тестовые прогоны со статистикой';

-- 12. Представление статистики по тест-кейсам
CREATE OR REPLACE VIEW v_test_case_statistics AS
SELECT 
    tc.test_case_id,
    tc.name,
    p.name AS priority,
    u.full_name AS owner,
    tc.is_automated,
    tc.is_active,
    COUNT(DISTINCT tres.test_result_id) AS total_executions,
    COUNT(DISTINCT CASE WHEN s.code = 'PASSED' THEN tres.test_result_id END) AS passed_count,
    COUNT(DISTINCT CASE WHEN s.code = 'FAILED' THEN tres.test_result_id END) AS failed_count,
    ROUND(
        100.0 * COUNT(DISTINCT CASE WHEN s.code = 'PASSED' THEN tres.test_result_id END) / 
        NULLIF(COUNT(DISTINCT tres.test_result_id), 0), 
        2
    ) AS pass_rate_percent,
    ROUND(AVG(tres.execution_duration_minutes), 2) AS avg_duration_minutes
FROM test_cases tc
JOIN priorities p ON tc.priority_id = p.priority_id
JOIN users u ON tc.owner_user_id = u.user_id
LEFT JOIN test_run_items tri ON tc.test_case_id = tri.test_case_id
LEFT JOIN test_results tres ON tri.run_item_id = tres.run_item_id
LEFT JOIN statuses s ON tres.status_id = s.status_id
WHERE tc.is_active = TRUE
GROUP BY tc.test_case_id, tc.name, p.name, u.full_name, tc.is_automated, tc.is_active;

COMMENT ON VIEW v_test_case_statistics IS 'Статистика эффективности тест-кейсов';

-- 13. Представление сводки выполнения тестов
CREATE OR REPLACE VIEW v_test_execution_summary AS
SELECT
    trun.test_run_id,
    trun.name AS test_run_name,
    tp.name AS test_plan_name,
    av.version_string,
    COUNT(tri.test_case_id) AS total_cases,
    COUNT(CASE WHEN s.code = 'PASSED' THEN 1 END) AS passed,
    COUNT(CASE WHEN s.code = 'FAILED' THEN 1 END) AS failed,
    COUNT(CASE WHEN s.code = 'BLOCKED' THEN 1 END) AS blocked,
    COUNT(CASE WHEN s.code = 'SKIPPED' THEN 1 END) AS skipped,
    ROUND(
        100.0 * COUNT(CASE WHEN s.code = 'PASSED' THEN 1 END) / 
        NULLIF(COUNT(tri.test_case_id), 0), 
        2
    ) AS pass_rate,
    trun.start_date,
    trun.end_date,
    trun.status AS run_status
FROM test_runs trun
JOIN test_plans tp ON trun.test_plan_id = tp.test_plan_id
JOIN app_versions av ON trun.version_id = av.version_id
LEFT JOIN test_run_items tri ON trun.test_run_id = tri.test_run_id
LEFT JOIN test_results tr ON tri.run_item_id = tr.run_item_id
LEFT JOIN statuses s ON tr.status_id = s.status_id
GROUP BY trun.test_run_id, trun.name, tp.name, av.version_string, 
         trun.start_date, trun.end_date, trun.status;

COMMENT ON VIEW v_test_execution_summary IS 'Сводка выполнения тестов по прогонам';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_public_test_plans;
DROP VIEW IF EXISTS v_tester_assigned_cases;
DROP VIEW IF EXISTS v_lead_all_cases;
DROP VIEW IF EXISTS v_engineer_autotests;
DROP VIEW IF EXISTS v_shared_environments;
DROP VIEW IF EXISTS v_tester_my_results;
DROP VIEW IF EXISTS v_lead_all_results;
DROP VIEW IF EXISTS v_public_results;
DROP VIEW IF EXISTS v_shared_reports;
DROP VIEW IF EXISTS v_admin_users_and_roles;
DROP VIEW IF EXISTS v_active_test_runs;
DROP VIEW IF EXISTS v_test_case_statistics;
DROP VIEW IF EXISTS v_test_execution_summary;
-- +goose StatementEnd
