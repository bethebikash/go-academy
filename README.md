# Go Academy

Go Academy is a multi-user, interactive learning platform focused on teaching the Go programming language. It combines structured learning paths, detailed lessons, personal notes, coding challenges, isolated code execution, and learning analytics.

## Architecture

Built as a **modular monolith**:
- **Backend:** Go, Chi HTTP router, PostgreSQL, pgx/v5, SQLC, golang-migrate, Argon2id, JWT + rotating refresh tokens.
- **Frontend:** React, TypeScript, Vite, Tailwind CSS, TanStack Query, React Hook Form, Zod.
- **Database:** PostgreSQL 16+.
- **Execution Engine (Phase 5):** Isolated worker running code in restricted Docker containers (`--network none`, non-root, read-only rootfs, resource limits).

## Quick Start (Local Development)

### Prerequisites
- Go 1.27+
- Node.js 22+ & npm 10+
- Docker & Docker Compose
- PostgreSQL 16 (or run via Docker Compose)

### Environment Configuration
Copy the example environment file:
```bash
cp .env.example .env
```

### Running Backend
```bash
cd backend
go run cmd/api/main.go
```

### Running Frontend
```bash
cd frontend
npm install
npm run dev
```

## Documentation
- [Product Specification](docs/product-specification.md)
- [System Architecture](docs/architecture.md)
- [Database Schema & ERD](docs/database.md)
- [REST API Contracts](docs/api.md)
- [Security Architecture](docs/security.md)
- [Milestone Plan](docs/milestone-plan.md)
