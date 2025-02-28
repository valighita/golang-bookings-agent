BIN=bookings-ai-chat

all: build

build:
	go build -o $(BIN) cmd/main.go

build-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BIN) cmd/main.go

run: build
	go run cmd/main.go

build-docker:
	docker build -t $(BIN) .

run-docker: build-docker
	docker run -p 5001:8080 $(BIN)

clean:
	rm -f $(BIN)

.PHONY: all run clean
