-- +goose Up
-- +goose StatementBegin
-- Прямой доступ к таблицам (без представлений для автоматизации)
GRANT SELECT ON autotest_versions, env_configs TO db_cicd_system;

GRANT SELECT, INSERT, UPDATE ON test_runs, test_run_items,
    test_results, test_result_artifacts TO db_cicd_system;

GRANT SELECT, INSERT ON reports TO db_cicd_system;

GRANT SELECT ON test_cases, test_plans, priorities, statuses,
    tool_configs, app_versions TO db_cicd_system;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_cicd_system;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON autotest_versions, env_configs FROM db_cicd_system;
REVOKE SELECT, INSERT, UPDATE ON test_runs, test_run_items, test_results, test_result_artifacts FROM db_cicd_system;
REVOKE SELECT, INSERT ON reports FROM db_cicd_system;
REVOKE SELECT ON test_cases, test_plans, priorities, statuses, tool_configs, app_versions FROM db_cicd_system;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM db_cicd_system;
-- +goose StatementEnd
