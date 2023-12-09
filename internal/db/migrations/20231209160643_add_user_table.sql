-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE "user" (
  id uuid PRIMARY KEY,
  username varchar NOT NULL,
  email varchar NOT NULL,
  password varchar NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE "user";
-- +goose StatementEnd
