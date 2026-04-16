-- +goose Up
-- +goose StatementBegin
-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testcases_updated_at
BEFORE UPDATE ON test_cases
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testplans_updated_at
BEFORE UPDATE ON test_plans
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_envconfigs_updated_at
BEFORE UPDATE ON env_configs
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_toolconfigs_updated_at
BEFORE UPDATE ON tool_configs
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_autotests_updated_at
BEFORE UPDATE ON autotests
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

-------------------------------------------------------------------------------
-- Триггеры аудита — универсальная функция audit_generic() (ПолИБ-6.2)
-- Покрывают все активы A1–A7
-------------------------------------------------------------------------------

-- A7: users
CREATE TRIGGER trg_users_audit
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION audit_generic('user_id');

-- A7: user_roles (PK составной, используем user_id как ключ записи)
CREATE TRIGGER trg_user_roles_audit
AFTER INSERT OR UPDATE OR DELETE ON user_roles
FOR EACH ROW EXECUTE FUNCTION audit_generic('user_id');

-- A1: test_plans
CREATE TRIGGER trg_test_plans_audit
AFTER INSERT OR UPDATE OR DELETE ON test_plans
FOR EACH ROW EXECUTE FUNCTION audit_generic('test_plan_id');

-- A2: test_cases
CREATE TRIGGER trg_test_cases_audit
AFTER INSERT OR UPDATE OR DELETE ON test_cases
FOR EACH ROW EXECUTE FUNCTION audit_generic('test_case_id');

-- A3: autotests
CREATE TRIGGER trg_autotests_audit
AFTER INSERT OR UPDATE OR DELETE ON autotests
FOR EACH ROW EXECUTE FUNCTION audit_generic('autotest_id');

-- A3: autotest_versions
CREATE TRIGGER trg_autotest_versions_audit
AFTER INSERT OR UPDATE OR DELETE ON autotest_versions
FOR EACH ROW EXECUTE FUNCTION audit_generic('version_id');

-- A4: test_results
CREATE TRIGGER trg_test_results_audit
AFTER INSERT OR UPDATE OR DELETE ON test_results
FOR EACH ROW EXECUTE FUNCTION audit_generic('test_result_id');

-- A4: test_result_artifacts
CREATE TRIGGER trg_test_result_artifacts_audit
AFTER INSERT OR UPDATE OR DELETE ON test_result_artifacts
FOR EACH ROW EXECUTE FUNCTION audit_generic('artifact_id');

-- A5: env_configs
CREATE TRIGGER trg_env_configs_audit
AFTER INSERT OR UPDATE OR DELETE ON env_configs
FOR EACH ROW EXECUTE FUNCTION audit_generic('env_config_id');

-- A6: tool_configs
CREATE TRIGGER trg_tool_configs_audit
AFTER INSERT OR UPDATE OR DELETE ON tool_configs
FOR EACH ROW EXECUTE FUNCTION audit_generic('tool_config_id');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_updated_at         ON users;
DROP TRIGGER IF EXISTS trg_testcases_updated_at     ON test_cases;
DROP TRIGGER IF EXISTS trg_testplans_updated_at     ON test_plans;
DROP TRIGGER IF EXISTS trg_envconfigs_updated_at    ON env_configs;
DROP TRIGGER IF EXISTS trg_toolconfigs_updated_at   ON tool_configs;
DROP TRIGGER IF EXISTS trg_autotests_updated_at     ON autotests;

DROP TRIGGER IF EXISTS trg_users_audit                  ON users;
DROP TRIGGER IF EXISTS trg_user_roles_audit             ON user_roles;
DROP TRIGGER IF EXISTS trg_test_plans_audit             ON test_plans;
DROP TRIGGER IF EXISTS trg_test_cases_audit             ON test_cases;
DROP TRIGGER IF EXISTS trg_autotests_audit              ON autotests;
DROP TRIGGER IF EXISTS trg_autotest_versions_audit      ON autotest_versions;
DROP TRIGGER IF EXISTS trg_test_results_audit           ON test_results;
DROP TRIGGER IF EXISTS trg_test_result_artifacts_audit  ON test_result_artifacts;
DROP TRIGGER IF EXISTS trg_env_configs_audit            ON env_configs;
DROP TRIGGER IF EXISTS trg_tool_configs_audit           ON tool_configs;
-- +goose StatementEnd
