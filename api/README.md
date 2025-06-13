# ⚡ Workflow API

A Go-based API for managing and executing workflow automations. Provides endpoints to retrieve workflow definitions and execute workflows, with PostgreSQL for persistent storage.

## 🛠️ Tech Stack

- Go 1.23+
- PostgreSQL
- Docker (for development and deployment)

## 🚀 Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL
- Docker & Docker Compose (recommended for development)

### 1. Configure Database

Set the `DATABASE_URL` environment variable:

```
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=disable
```

Ensure PostgreSQL is running and accessible.

### 2. Run the API

- With Docker Compose (recommended):
  ```bash
  docker-compose up --build api
  ```
- Or run locally:
  ```bash
  go run main.go
  ```

## 📋 API Endpoints

| Method | Endpoint                         | Description                        |
| ------ | -------------------------------- | ---------------------------------- |
| GET    | `/api/v1/workflows/{id}`         | Load a workflow definition         |
| POST   | `/api/v1/workflows/{id}/execute` | Execute the workflow synchronously |

### Example Usage

#### GET workflow definition

```bash
curl http://localhost:8086/api/v1/workflows/weather-alert-workflow
```

#### POST execute workflow

```bash
curl -X POST http://localhost:8086/api/v1/workflows/weather-alert-workflow/execute \
     -H "Content-Type: application/json" \
     -d '{}'
```

## 🗄️ Database

- The API uses `api/pkg/db.DefaultConfig()` and reads the URI from `DATABASE_URL`.
- For schema/configuration details, see the main project README or this file's comments.
