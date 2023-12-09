NAME=rtm

run:
	go run cmd/rtm/rtm.go

build:
	go build -o bin/$(NAME) cmd/rtm/rtm.go

start:
	./bin/$(NAME)

clean:
	rm -rf bin

.PHONY: run build start clean
