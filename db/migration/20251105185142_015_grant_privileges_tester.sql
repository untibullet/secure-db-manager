-- +goose Up
-- +goose StatementBegin
-- Доступ к представлениям
GRANT SELECT ON v_public_test_plans, v_tester_assigned_cases,
    v_shared_environments, v_tester_my_results, v_shared_reports TO db_tester;

-- Доступ к тест-кейсам (только назначенным)
GRANT SELECT, INSERT, UPDATE ON test_cases, test_case_steps TO db_tester;

-- Доступ к результатам
GRANT SELECT, INSERT, UPDATE ON test_results, test_result_artifacts TO db_tester;

-- Чтение справочников
GRANT SELECT ON test_plans, test_runs, test_run_items, priorities,
    statuses, users, roles, env_configs, tool_configs TO db_tester;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_tester;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON v_public_test_plans, v_tester_assigned_cases, v_shared_environments, v_tester_my_results, v_shared_reports FROM db_tester;
REVOKE SELECT, INSERT, UPDATE ON test_cases, test_case_steps FROM db_tester;
REVOKE SELECT, INSERT, UPDATE ON test_results, test_result_artifacts FROM db_tester;
REVOKE SELECT ON test_plans, test_runs, test_run_items, priorities, statuses, users, roles, env_configs, tool_configs FROM db_tester;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM db_tester;
-- +goose StatementEnd

