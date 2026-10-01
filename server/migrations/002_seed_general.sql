-- +goose Up
-- +goose StatementBegin
INSERT INTO groups (id, name, created_at) VALUES ('general', 'General', '2026-10-01T00:00:00Z') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM groups WHERE id = 'general';
-- +goose StatementEnd
