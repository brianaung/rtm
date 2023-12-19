-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE room (
  id uuid PRIMARY KEY,
  roomname varchar NOT NULL,
  creator_id uuid NOT NULL,
  CONSTRAINT fk_creator
    FOREIGN KEY(creator_id)
      REFERENCES "user"(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE room;
-- +goose StatementEnd
