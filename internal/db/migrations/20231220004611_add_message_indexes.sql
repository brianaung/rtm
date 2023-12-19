-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE INDEX message_room_id_fkey ON message(room_id);
CREATE INDEX message_user_id_fkey ON message(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS message_room_id_fkey;
DROP INDEX IF EXISTS message_user_id_fkey;
-- +goose StatementEnd
