include .env

NAME=rtm
RTM_PATH=cmd/rtm/main.go

run: templ
	go run $(RTM_PATH)

build:
	go build -o dist/$(NAME) $(RTM_PATH)

start: build
	./dist/$(NAME)

clean:
	rm -rf dist

templ: tailwind
	templ generate

tailwind:
	npx tailwindcss -i ./view/input.css -o ./dist/output.css

up:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" up

down:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" down

status:
	@goose -dir internal/db/migrations/ postgres "user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" status

.PHONY: run build start clean templ tailwind up down status 
