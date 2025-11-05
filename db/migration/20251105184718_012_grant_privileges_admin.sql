-- +goose Up
-- +goose StatementBegin
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO db_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO db_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO db_admin;
GRANT ALL PRIVILEGES ON DATABASE postgres TO db_admin;

-- Полный доступ к представлениям
GRANT SELECT, INSERT, UPDATE, DELETE ON v_admin_users_and_roles TO db_admin;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM db_admin;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM db_admin;
REVOKE EXECUTE ON ALL FUNCTIONS IN SCHEMA public FROM db_admin;
REVOKE ALL PRIVILEGES ON DATABASE postgres FROM db_admin;
REVOKE SELECT, INSERT, UPDATE, DELETE ON v_admin_users_and_roles FROM db_admin;
-- +goose StatementEnd

