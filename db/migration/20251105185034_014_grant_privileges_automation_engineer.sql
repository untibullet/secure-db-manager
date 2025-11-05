-- +goose Up
-- +goose StatementBegin
-- Доступ к представлениям
GRANT SELECT ON v_public_test_plans, v_lead_all_cases, v_engineer_autotests,
    v_shared_environments, v_lead_all_results, v_shared_reports,
    v_active_test_runs, v_test_case_statistics TO db_automation_engineer;

-- Доступ к автотестам
GRANT SELECT, INSERT, UPDATE, DELETE ON autotests, autotest_versions TO db_automation_engineer;

-- Доступ к окружениям
GRANT SELECT, INSERT, UPDATE ON env_configs, env_config_params TO db_automation_engineer;

-- Доступ к результатам и тест-кейсам
GRANT SELECT ON test_plans, test_cases, test_runs, test_results,
    test_result_artifacts, priorities, statuses, users TO db_automation_engineer;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_automation_engineer;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON v_public_test_plans, v_lead_all_cases, v_engineer_autotests, v_shared_environments, v_lead_all_results, v_shared_reports, v_active_test_runs, v_test_case_statistics FROM db_automation_engineer;
REVOKE SELECT, INSERT, UPDATE, DELETE ON autotests, autotest_versions FROM db_automation_engineer;
REVOKE SELECT, INSERT, UPDATE ON env_configs, env_config_params FROM db_automation_engineer;
REVOKE SELECT ON test_plans, test_cases, test_runs, test_results, test_result_artifacts, priorities, statuses, users FROM db_automation_engineer;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM db_automation_engineer;
-- +goose StatementEnd

