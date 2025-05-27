.PHONY: build run test clean serve swagger swagger-script

# Build variables
BINARY_NAME=gitclient
MAIN_PATH=./cmd/gitclient

# Build the application
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)/main.go $(ARGS)

# Run the API server
serve:
	go run $(MAIN_PATH)/main.go -command=serve -port=$(PORT)

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Generate Swagger documentation using swag CLI
swagger:
	swag init -g pkg/api/server.go -o docs

# Generate Swagger documentation using Go script
swagger-script:
	go run cmd/swagger/main.go

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Default target
all: deps fmt vet test build