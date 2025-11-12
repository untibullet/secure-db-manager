-- +goose Up
-- +goose StatementBegin
-- Вставка базовых приоритетов
INSERT INTO priorities (code, name, rank, color_hex) VALUES
    ('CRITICAL', 'Критический', 5, '#FF0000'),
    ('HIGH', 'Высокий', 4, '#FF6600'),
    ('MEDIUM', 'Средний', 3, '#FFCC00'),
    ('LOW', 'Низкий', 2, '#00CC00'),
    ('TRIVIAL', 'Незначительный', 1, '#CCCCCC');

-- Вставка базовых статусов
INSERT INTO statuses (code, name, category, is_final) VALUES
    ('PASSED', 'Пройден', 'TEST_EXECUTION', TRUE),
    ('FAILED', 'Провален', 'TEST_EXECUTION', TRUE),
    ('BLOCKED', 'Заблокирован', 'TEST_EXECUTION', TRUE),
    ('SKIPPED', 'Пропущен', 'TEST_EXECUTION', TRUE),
    ('IN_PROGRESS', 'В процессе', 'TEST_EXECUTION', FALSE),
    ('NOT_RUN', 'Не запущен', 'TEST_EXECUTION', FALSE),
    ('DRAFT', 'Черновик', 'PLAN', FALSE),
    ('ACTIVE', 'Активный', 'PLAN', FALSE),
    ('COMPLETED', 'Завершен', 'PLAN', TRUE),
    ('ARCHIVED', 'Архивирован', 'PLAN', TRUE);

-- Вставка базовых ролей
INSERT INTO roles (code, name, description, access_level) VALUES
    ('ADMIN', 'Администратор', 'Полный доступ к системе, управление пользователями и ролями', 10),
    ('TEST_LEAD', 'Тест-лид', 'Управление тест-планами, кейсами и командой тестирования', 8),
    ('AUTOMATION_ENGINEER', 'Инженер по автоматизации', 'Разработка и поддержка автотестов', 7),
    ('TESTER', 'Тестировщик', 'Выполнение тестов и создание тест-кейсов', 5),
    ('CI_CD_SYSTEM', 'CI/CD система', 'Автоматическое выполнение тестов через CI/CD pipeline', 6),
    ('DEVELOPER', 'Разработчик', 'Просмотр результатов и создание конфигураций', 4),
    ('GUEST', 'Гость', 'Только просмотр публичных данных', 1);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
TRUNCATE TABLE roles, statuses, priorities RESTART IDENTITY CASCADE;
-- +goose StatementEnd
