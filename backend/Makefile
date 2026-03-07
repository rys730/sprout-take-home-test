.PHONY: swagger build run test clean tidy sqlc

# Generate sqlc code from queries
sqlc:
	sqlc generate

# Generate Swagger docs from annotations
swagger:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Build the server binary
build: sqlc swagger
	go build -o bin/server cmd/server/main.go

# Run the server
run: swagger
	go run cmd/server/main.go

# Run tests
test:
	go test ./... -v

# Clean generated files
clean:
	rm -rf docs/ bin/server

# Tidy modules
tidy:
	go mod tidy

# Run migrations only
migrate:
	go run cmd/server/main.go -migrate-only 2>/dev/null || true

.PHONY: migrate-db-new
migrate-db-new: ## Create a new sql db migration
ifndef file_name
	$(error file_name is not defined. Use make migrate-db-new file_name=<name>)
endif
	goose -dir migrations create $(file_name) sql