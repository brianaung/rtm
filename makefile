include .env

NAME=rtm
RTM_PATH=cmd/rtm/main.go

run: templ
	go run $(RTM_PATH)

templ:
	templ generate

build:
	go build -o dist/$(NAME) $(RTM_PATH)

start:
	./dist/$(NAME)

clean:
	rm -rf dist

tailwind:
	npx tailwindcss -i ./view/input.css -o ./dist/output.css --watch

up:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" up

down:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" down

status:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" status

.PHONY: run build start clean up down status tailwind
