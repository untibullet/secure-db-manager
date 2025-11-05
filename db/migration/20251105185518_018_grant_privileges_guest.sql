-- +goose Up
-- +goose StatementBegin
-- Только чтение публичных данных
GRANT SELECT ON v_public_test_plans, v_public_results, v_shared_reports TO db_guest;

GRANT SELECT ON priorities, statuses TO db_guest;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON v_public_test_plans, v_public_results, v_shared_reports FROM db_guest;
REVOKE SELECT ON priorities, statuses FROM db_guest;
-- +goose StatementEnd
