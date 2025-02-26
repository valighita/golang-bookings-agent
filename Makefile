BIN=main

all: build

build: $(BIN)
	go build -o $(BIN) cmd/main.go

run: build
	go run cmd/main.go

clean:
	rm -f $(BIN)

.PHONY: all run clean
