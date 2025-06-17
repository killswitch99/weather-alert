# ⚡ Workflow Editor

A modern workflow editor app for designing and executing custom automation workflows (e.g., weather notifications). Users can visually build workflows, configure parameters, and view real-time execution results.

## 🛠️ Tech Stack

- **Frontend:** React + TypeScript, @xyflow/react (drag-and-drop), Radix UI, Tailwind CSS, Vite
- **Backend:** Go API, PostgreSQL database
- **DevOps:** Docker Compose for orchestration, hot reloading for rapid development

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose (recommended for development)
- Node.js v18+ (for local frontend development)
- Go v1.23+ (for local backend development)

> **Tip:** Node.js and Go are only required if you want to run frontend or backend outside Docker.

### 1. Start All Services

```bash
docker-compose up --build
```

- This launches frontend, backend, and database with hot reloading enabled for code changes.
- To stop and clean up:
  ```bash
  docker-compose down
  ```

### 2. Access Applications

- **Frontend (Workflow Editor):** [http://localhost:3003](http://localhost:3003)
- **Backend API:** [http://localhost:8086](http://localhost:8086)
- **Database:** PostgreSQL on `localhost:5876`

### 3. Verify Setup

1. Open [http://localhost:3003](http://localhost:3003) in your browser.
2. You should see the workflow editor with sample nodes.

## 🏗️ Project Architecture

```text
workflow-challenge-v2/
├── api/                    # Go Backend (Port 8086)
│   ├── cmd/api/            # Main entrypoint for API server
│   ├── internal/
│   │   ├── service/        # Execution engine, node executors, business logic
│   │   ├── repository/     # Persistence layer (DB access)
│   │   ├── workflow/       # API handlers
│   │   └── integration/    # Integration tests
│   ├── pkg/models/         # Shared Go types and models
│   ├── scripts/            # DB migration and init scripts
│   ├── migrations/         # SQL migration files
│   ├── go.mod
│   ├── Dockerfile
│   └── README.md           # Backend API docs
├── web/                    # React Frontend (Port 3003)
│   ├── src/
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml      # Orchestrates DB, API, and frontend
└── README.md               # Project overview (this file)
```

## 🔧 Development Workflow

### 🌐 Frontend

- Edit files in `web/src/` and see changes instantly at [http://localhost:3003](http://localhost:3003) (hot reloading via Vite).

### 🖥️ Backend

- Edit files in `api/` and changes are reflected automatically (hot reloading in Docker).
- If you add new dependencies or make significant changes, rebuild the API container:
  ```bash
  docker-compose up --build api
  ```

### 🗄️ Database

- Schema/configuration details: see [API README](api/README.md#database)
- After schema changes or migrations, restart the database:
  ```bash
  docker-compose restart postgres
  ```
- To apply schema changes to the API after updating the database:
  ```bash
  docker-compose restart api
  ```