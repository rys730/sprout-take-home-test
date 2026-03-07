# Sprout Take Home Test - Backend

## Overview

This is a Golang backend service following clean architecture principles, built with Echo framework, PostgreSQL, and sqlc.

## Architecture

The project follows clean architecture with three main layers:

- **Handler Layer**: HTTP controllers handling requests and responses
- **Use Case Layer**: Business logic implementation
- **Repository Layer**: Data access using sqlc-generated queries

## Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/                    # Configuration management
│   ├── domain/                    # Domain entities and interfaces
│   ├── handler/                   # HTTP handlers and middleware
│   ├── repository/                # Data access layer
│   ├── usecase/                   # Business logic layer
│   └── infrastructure/           # External services (db, logger, auth)
├── db/
│   └── queries/                   # sqlc generated files
├── migrations/                    # Database migrations
├── .env.example                   # Environment variables template
├── sqlc.yaml                      # sqlc configuration
├── Dockerfile                     # Docker image build
└── go.mod                         # Go module definition
```

## Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Local Development

1. **Copy environment variables:**
    ```bash
    cp backend/.env.example backend/.env
    ```

2. **Install dependencies:**
    ```bash
    cd backend
    go mod download
    ```

3. **Start PostgreSQL:**
    ```bash
    docker-compose up -d postgres
    ```

4. **Generate sqlc queries:**
    ```bash
    sqlc generate
    ```

5. **Run server:**
    ```bash
    go run cmd/server/main.go
    ```

**Note:** Database migrations run automatically on startup using goose.

2. **Install dependencies:**
   ```bash
   cd backend
   go mod download
   ```

3. **Start PostgreSQL:**
   ```bash
   docker-compose up -d postgres
   ```

4. **Run migrations (if you have any):**
   ```bash
   # Add your migration tool here (e.g., migrate, goose, etc.)
   ```

5. **Generate sqlc queries:**
   ```bash
   sqlc generate
   ```

6. **Run the server:**
   ```bash
   go run cmd/server/main.go
   ```

### Docker Compose

1. **Start all services:**
   ```bash
   docker-compose up -d
   ```

2. **View logs:**
   ```bash
   docker-compose logs -f backend
   ```

3. **Stop services:**
   ```bash
   docker-compose down
   ```

## API Endpoints

### Health Check
- `GET /health` - Health check endpoint

### Authentication
- `POST /api/v1/auth/login` - Login and get JWT token
- `POST /api/v1/auth/refresh` - Refresh JWT token

### Protected Routes
Protected routes require a valid JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

## Configuration

Configuration is loaded from environment variables. See `.env.example` for available options:

- **Database**: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
- **Server**: SERVER_PORT, SERVER_HOST
- **JWT**: JWT_SECRET, JWT_EXPIRATION_HOURS
- **CORS**: CORS_ALLOWED_ORIGINS, CORS_ALLOWED_METHODS, CORS_ALLOWED_HEADERS
- **Environment**: ENV (development/production)

## Development

### Adding New Features

1. **Define domain entities** in `internal/domain/entities.go`
2. **Create repository interface** in `internal/repository/interface.go`
3. **Implement repository** in `internal/repository/postgres.go`
4. **Create use case** in `internal/usecase/interface.go`
5. **Create handler** in `internal/handler/`
6. **Register routes** in `cmd/server/main.go`

### Database Migrations with Goose

Migrations are managed using goose and run automatically when the server starts.

**To create a new migration:**
```bash
goose -dir migrations create name_of_migration sql
```

**To manually run migrations:**
```bash
goose -dir migrations postgres "user=postgres dbname=sprout_db sslmode=disable" up
```

**To roll back the last migration:**
```bash
goose -dir migrations postgres "user=postgres dbname=sprout_db sslmode=disable" down
```

**Migration files:**
- Migration files are located in the `migrations/` directory
- Follow the naming convention: `XXXXX_description.sql`
- Use `-- +goose Up` for up migrations
- Use `-- +goose Down` for down migrations

### Database Queries with sqlc

1. Create SQL files in `db/queries/`
2. Run `sqlc generate` to generate type-safe Go code
3. Use generated queries in your repository layer

### Running Tests

```bash
go test ./...
```

## Building

```bash
go build -o bin/server ./cmd/server
```

## Dependencies

- Echo v4 - Web framework
- pgx/v5 - PostgreSQL driver
- sqlc - SQL query generator
- Goose - Database migration tool
- Viper - Configuration management
- zerolog - Structured logging
- golang-jwt/jwt - JWT authentication

## Security Notes

- Change `JWT_SECRET` in production
- Use strong database passwords
- Enable SSL for database connections in production
- Review and update CORS settings
- Use HTTPS in production

## TODO

- Add database migrations (migrate, goose, or similar)
- Implement actual user authentication logic
- Add more comprehensive error handling
- Add API documentation (Swagger/OpenAPI)
- Add integration tests
- Implement request validation
- Add rate limiting middleware
