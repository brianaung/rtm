-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE room_user (
    room_id uuid,
    user_id uuid,
    PRIMARY KEY(room_id, user_id),
    CONSTRAINT fk_room FOREIGN KEY(room_id) REFERENCES room(id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES "user"(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE room_user;
-- +goose StatementEnd
