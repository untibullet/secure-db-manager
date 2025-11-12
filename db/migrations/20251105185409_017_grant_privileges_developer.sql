-- +goose Up
-- +goose StatementBegin
-- Просмотр результатов
GRANT SELECT ON v_public_test_plans, v_lead_all_results,
    v_shared_reports, v_active_test_runs, v_test_execution_summary TO db_developer;

-- Управление конфигурациями
GRANT SELECT, INSERT, UPDATE ON env_configs, env_config_params,
    tool_configs, tool_config_params, config_errors TO db_developer;

GRANT SELECT, INSERT ON app_versions TO db_developer;

-- Чтение данных
GRANT SELECT ON test_plans, test_cases, test_runs, test_results,
    test_result_artifacts, priorities, statuses, users TO db_developer;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_developer;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON v_public_test_plans, v_lead_all_results, v_shared_reports, v_active_test_runs, v_test_execution_summary FROM db_developer;
REVOKE SELECT, INSERT, UPDATE ON env_configs, env_config_params, tool_configs, tool_config_params, config_errors FROM db_developer;
REVOKE SELECT, INSERT ON app_versions FROM db_developer;
REVOKE SELECT ON test_plans, test_cases, test_runs, test_results, test_result_artifacts, priorities, statuses, users FROM db_developer;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM db_developer;
-- +goose StatementEnd
