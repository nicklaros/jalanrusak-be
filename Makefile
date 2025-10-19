.PHONY: help build run test fmt tidy swagger clean

BINARY := bin/server
SWAGGER_MAIN := cmd/server/main.go
SWAGGER_OUT := docs

help:
	@echo "Available targets:"
	@echo "  make build    # compile the HTTP server binary"
	@echo "  make run      # run the HTTP server"
	@echo "  make test     # execute go test ./..."
	@echo "  make fmt      # format all Go source files"
	@echo "  make tidy     # tidy go.mod and go.sum"
	@echo "  make swagger  # regenerate Swagger documentation"
	@echo "  make clean    # remove build artifacts"

build:
	@go build -o $(BINARY) $(SWAGGER_MAIN)

run:
	@go run $(SWAGGER_MAIN)

test:
	@go test ./...


fmt:
	@go fmt ./...

tidy:
	@go mod tidy

swagger:
	@command -v swag >/dev/null || { echo "swag CLI not found. Install with 'go install github.com/swaggo/swag/cmd/swag@latest'"; exit 1; }
	@swag init -g $(SWAGGER_MAIN) -o $(SWAGGER_OUT)

clean:
	@rm -f $(BINARY)
