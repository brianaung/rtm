include .env

NAME=rtm
RTM_PATH=cmd/rtm/main.go

run:
	go run $(RTM_PATH)

build:
	go build -o dist/$(NAME) $(RTM_PATH)

start:
	./dist/$(NAME)

clean:
	rm -rf dist

up:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" up

down:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" down

status:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" status

.PHONY: run build start clean up down status
