-- +goose Up
-- +goose StatementBegin
-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testcases_updated_at BEFORE UPDATE ON test_cases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testplans_updated_at BEFORE UPDATE ON test_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_envconfigs_updated_at BEFORE UPDATE ON env_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_toolconfigs_updated_at BEFORE UPDATE ON tool_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_autotests_updated_at BEFORE UPDATE ON autotests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Триггеры аудита для критических таблиц
CREATE TRIGGER trg_users_audit
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER trg_userroles_audit
    AFTER INSERT OR UPDATE OR DELETE ON user_roles
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER trg_testcases_audit
    AFTER INSERT OR UPDATE OR DELETE ON test_cases
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER trg_testplans_audit
    AFTER INSERT OR UPDATE OR DELETE ON test_plans
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TRIGGER IF EXISTS trg_testcases_updated_at ON test_cases;
DROP TRIGGER IF EXISTS trg_testplans_updated_at ON test_plans;
DROP TRIGGER IF EXISTS trg_envconfigs_updated_at ON env_configs;
DROP TRIGGER IF EXISTS trg_toolconfigs_updated_at ON tool_configs;
DROP TRIGGER IF EXISTS trg_autotests_updated_at ON autotests;
DROP TRIGGER IF EXISTS trg_users_audit ON users;
DROP TRIGGER IF EXISTS trg_userroles_audit ON user_roles;
DROP TRIGGER IF EXISTS trg_testcases_audit ON test_cases;
DROP TRIGGER IF EXISTS trg_testplans_audit ON test_plans;
-- +goose StatementEnd
