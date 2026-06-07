# Golang SaaS Starter Kit

A production-ready, highly-scalable Golang SaaS Backend Starter Kit built with clean architecture principles. It serves as a reusable foundation for building modern software-as-a-service applications.

## Tech Stack

- **Framework**: [Gin Gonic](https://github.com/gin-gonic/gin) (HTTP Router)
- **Database**: PostgreSQL with [GORM](https://gorm.io/) (Object Relational Mapper)
- **Configuration**: [Viper](https://github.com/spf13/viper) & [Godotenv](https://github.com/joho/godotenv)
- **Logging**: [Zerolog](https://github.com/rs/zerolog) (Structured JSON Logs)
- **Validation**: [Validator v10](https://github.com/go-playground/validator)
- **Security**: [Casbin](https://casbin.org/) (RBAC Authorization) & [JWT](https://github.com/golang-jwt/jwt)
- **Documentation**: [Swagger](https://swagger.io/) (via Swag)
- **Development**: Docker Compose & [Air](https://github.com/air-verse/air) (Hot reloading)
- **Testing**: [Testify](https://github.com/stretchr/testify)

---

## Directory Structure

```
├── cmd
│   ├── api          # Main application entry point
│   ├── migrate      # Database migrations CLI binary (Up, Rollback, Automigrate)
│   └── seed         # Database seeder binary
├── configs          # Environment configurations & Casbin policies
├── docs             # Auto-generated Swagger documentation
├── internal
│   ├── entities     # Global GORM entities/Base models
│   ├── middleware   # CORS, RequestID, Recovery, Logging, Auth, and Ownership middlewares
│   └── modules      # Business features
│       ├── audit    # Audit Trail module (reusable activity logger)
│       ├── health   # Health Check module
│       └── user     # User CRUD module (Clean Architecture example)
└── pkg
    ├── auth         # JWT and Casbin engine initializers
    ├── database     # GORM DB connection pool helper
    ├── logger       # Zerolog logger wrappers
    ├── response     # Reusable standard API response payloads
    └── validator    # Struct tag validation helpers
```

---

## Getting Started

### Prerequisites
Make sure you have the following installed on your machine:
- [Go](https://go.dev/) (v1.25 or later)
- [Docker](https://www.docker.com/) & Docker Compose

### 1. Setup Environment
Copy the example environment file:
```bash
cp .env.example .env
```

### 2. Launch Database (Docker)
Start the PostgreSQL container:
```bash
make docker-up
```
*Note: A PgAdmin instance is also spun up at `http://localhost:5050` (credentials: `admin@cellar.com` / `admin`).*

### 3. Run Database Migrations & Seeders
Create the database schema and populate it with initial seed users (Admin, Moderator, and Standard Users):
```bash
make automigrate
make seed
```

### 4. Start Development Server
Boot the API server with hot reloading enabled (Air):
```bash
make run
```
If you don't have `air` installed globally, the command automatically falls back to `go run cmd/api/main.go`.

---

## API Documentation (Swagger)

API routes are documented using Swagger annotations. To generate or update the documentation, run:
```bash
make swagger
```
Once the server is running, access the interactive Swagger UI at:
`http://localhost:8080/swagger/index.html`

---

## Core Features & Examples

### 1. Reusable Ownership Middleware
In [internal/middleware/ownership.go](file:///Users/syafrisurya/project/belajar/cellar-backend/internal/middleware/ownership.go), we provide a reusable middleware:
```go
middleware.AuthorizeOwnership("id", middleware.UserSelfProvider)
```
You can pass any custom resource resolver to look up other tables (e.g. Foods, Reservations) and enforce that the current user owns that item or bypass it if they are an `admin`.

### 2. Structured Audit Trails
Every mutation in the User module registers an activity trail in PostgreSQL (table `audit_trails`). Services can log events like:
```go
s.auditServ.Record(operatorID, "User", "Create", "Created user test@example.com")
```

### 3. Casbin RBAC Policies
Role-based route permissions are defined in [configs/rbac_policy.csv](file:///Users/syafrisurya/project/belajar/cellar-backend/configs/rbac_policy.csv). Modify this file to add/remove role access control paths dynamically.

---

## Development Tooling (Makefile)

Here is a summary of available make tasks:

| Command | Action |
|---|---|
| `make run` | Run dev server with hot reload |
| `make build` | Compile backend binary to `bin/api` |
| `make test` | Run unit tests (uses SQLite in-memory DB) |
| `make lint` | Run code formatter & compiler vet checks |
| `make swagger` | Rebuild Swagger documentation |
| `make seed` | Populate seed database users |
| `make migrate` | Execute versioned migrations |
| `make rollback` | Rollback database tables |
| `make automigrate` | GORM-driven auto schema synchronization |
| `make docker-up` | Boot local PostgreSQL & PgAdmin |
| `make docker-down` | Tear down local Docker containers |

---

## Testing

Unit tests are written using standard `testing`, `testify`, and GORM's SQLite in-memory driver.
To execute tests, run:
```bash
make test
```
 Mocks are generated and mapped manually or using mockery for the service layer logic testing.
