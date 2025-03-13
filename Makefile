BIN=bookings-ai-chat

all: build

build:
	go build -o $(BIN) cmd/main.go

run: build
	./$(BIN)

run-cli: build
	./$(BIN) cli

build-docker:
	docker build -t $(BIN) .

run-docker: build-docker
	docker run -p 5001:8080 $(BIN)

clean:
	rm -f $(BIN)

.PHONY: all build run run-cli build-docker run-docker clean
