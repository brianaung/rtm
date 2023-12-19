-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE message(
    id uuid PRIMARY KEY,
    msg text NOT NULL, 
    time timestamptz NOT NULL,
    room_id uuid NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT fk_room FOREIGN KEY(room_id) REFERENCES room(id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES "user"(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE message;
-- +goose StatementEnd
