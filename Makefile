APP_NAME := server
CMD_DIR  := ./cmd/server
BIN_DIR  := ./bin

.PHONY: build run clean test fmt vet

build:
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run: build
	$(BIN_DIR)/$(APP_NAME)

clean:
	rm -rf $(BIN_DIR)

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

docker:
	docker compose up --build
