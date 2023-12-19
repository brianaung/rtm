-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE INDEX message_time_idx ON message (time DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS message_time_idx
-- +goose StatementEnd
