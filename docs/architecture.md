# Go Academy --- System Architecture

**Document Version:** 1.0  
**Status:** Approved Architecture Baseline  
**Scope:** Modular Monolith Architecture & Technical Design  

---

## 1. High-Level Architecture Overview

Go Academy is architected as a **modular monolith** consisting of:
1. **Backend Service (`backend/`):** A single Go binary hosting the Chi HTTP REST API, domain modules, database persistence layer, and background tasks.
2. **Admin CLI (`backend/cmd/admin`):** An administrative command-line utility used for initial platform bootstrap and operational tasks without exposing public bootstrap endpoints.
3. **Frontend Application (`frontend/`):** A client-side Single Page Application (SPA) built with React, TypeScript, Vite, Tailwind CSS, and TanStack Query.
4. **Primary Database:** PostgreSQL 16+ as the single source of truth for identity, content, user notes, bookmarks, and submissions.
5. **Future Execution Boundary (Phase 5):** An asynchronous worker process that pulls execution jobs and runs learner Go code in isolated, resource-constrained, non-root Docker containers completely detached from the API process and database credentials.

```
+-------------------------------------------------------------------------------+
|                                Client Browser                                 |
|  +-------------------------------------------------------------------------+  |
|  |           React + TypeScript Single Page Application (Vite)             |  |
|  |  [Router: React Router] [Data: TanStack Query] [Forms: HookForm + Zod]   |  |
|  |  [Editor: Monaco] [Markdown: Safe AST Renderer] [Theme: Dark First]     |  |
|  +-------------------------------------------------------------------------+  |
+---------------------------------------|---------------------------------------+
                                        | HTTPS / JSON REST API (/api/v1)
                                        v
+-------------------------------------------------------------------------------+
|                       Go API Process (Modular Monolith)                       |
|                                                                               |
|  +-------------------------------------------------------------------------+  |
|  | Middleware Stack: RealIP, RequestID, Logger (slog), CORS, Recoverer,    |  |
|  | RateLimiter, BodyLimit, AuthVerifier, RBAC Enforcement                  |  |
|  +-------------------------------------------------------------------------+  |
|                                                                               |
|  +--------------------------- Domain Modules ----------------------------+  |
|  | [auth]          - Argon2id, JWT Access Tokens, Rotating Refresh Tokens  |  |
|  | [users]         - User Profiles, Activation & Deactivation             |  |
|  | [roles]         - RBAC Policy (Admin, Learner)                         |  |
|  | [invitations]   - Secure Tokens (SHA-256), Expiry, Atomic Registration  |  |
|  | [courses]       - Learning Paths & Course Hierarchies                  |  |
|  | [modules]       - Course Modules & Reordering                          |  |
|  | [lessons]       - Structured Markdown Content, Prerequisites, Drafts   |  |
|  | [notes]         - User Markdown Notes, Optimistic Locking, Autosave    |  |
|  | [bookmarks]     - Private Content Bookmarking                          |  |
|  | [challenges]    - Problem Statements, Hints, Starter Code, Test Cases   |  |
|  | [submissions]   - Code Submission Records & Stored History             |  |
|  | [progress]      - Idempotent Lesson Progress, Streaks, Analytics       |  |
|  | [admin]         - Administrative Aggregates & System Dashboards        |  |
|  +-------------------------------------------------------------------------+  |
|                                                                               |
|  +---------------------- Persistence & Infrastructure ---------------------+  |
|  | [database]      - pgxpool Connection Lifecycle, Health Ping             |  |
|  | [sqlc]          - Type-safe Generated SQL Repositories                 |  |
|  | [migrations]    - golang-migrate Versioned Schema Management           |  |
|  +-------------------------------------------------------------------------+  |
+---------------------------------------|---------------------------------------+
                                        | Parameterized SQL (pgx pool)
                                        v
+-------------------------------------------------------------------------------+
|                               PostgreSQL Database                             |
|  - Users & Roles              - Invitations & Sessions (Hashed secrets)       |
|  - Learning Paths & Lessons   - Notes, Tags & Bookmarks                       |
|  - Challenges & Test Cases    - Submissions & Execution Jobs                  |
+-------------------------------------------------------------------------------+

+-------------------------------------------------------------------------------+
|                  Isolated Execution Boundary (Phase 5)                        |
|                                                                               |
|  +--------------------------+         Polls execution jobs via DB/Worker      |
|  |  Execution Runner Worker | <===========================================+   |
|  +-------------+------------+                                                 |
|                | Spawns ephemeral container                                   |
|                v                                                              |
|  +-------------------------------------------------------------------------+  |
|  | Restricted Docker Container                                             |  |
|  | - Non-root user (uid 10001)   - Read-only rootfs with tmpfs /tmp        |  |
|  | - No Docker socket mounted    - No network access (`--network none`)    |  |
|  | - CPU quota (1 core)          - Memory limit (128 MB)                   |  |
|  | - Process limit (pids=30)     - Wall-clock timeout (5 seconds)          |  |
|  | - Bounded output buffer (64KB)- Drops all Linux capabilities (cap-drop) |  |
|  +-------------------------------------------------------------------------+  |
+-------------------------------------------------------------------------------+
```

---

## 2. Backend Architecture & Package Structure

### 2.1 Layered Module Architecture
Each domain module within `backend/internal/<module>` follows a strict separation of concerns:
```
backend/internal/<module>/
├── handler.go     # HTTP transport layer: parses JSON, invokes service, writes envelopes
├── service.go     # Core business logic, validation, orchestrates transactions
├── repository.go  # Data access interface wrapping SQLC-generated queries
└── model.go       # Domain-specific structs, DTOs, and error types
```

### 2.2 Architectural Rules & Invariants
1. **Handlers contain no business logic:** Handlers exclusively parse HTTP requests, perform structural schema validation, invoke service methods, and serialize JSON responses.
2. **Explicit Interfaces:** Cross-module communication occurs strictly via defined Go interfaces (e.g. `users.Service` or `invitations.Validator`), never by reaching into another module's internal repository or unexported helpers.
3. **Database Access via SQLC:** All database interactions are written in plain, version-controlled `.sql` query files and compiled into type-safe Go structs using SQLC. Dynamic SQL generation or uncontrolled ORMs are prohibited.
4. **Transaction Boundaries:** Transactions are managed at the service layer via pgx transactional handles passed into SQLC's `WithTx(tx)`.
5. **No Microservices:** The application is compiled into a single executable (`backend/cmd/api/main.go`). Separate deployment artifacts are restricted to the Admin CLI (`backend/cmd/admin/main.go`) and the future isolated runner worker (`backend/cmd/runner/main.go`).

### 2.3 Monorepo Backend Tree
```
backend/
├── cmd/
│   ├── api/
│   │   └── main.go          # Application entrypoint: wire dependencies & start HTTP server
│   └── admin/
│       └── main.go          # Bootstrap CLI: seed initial admin account safely
├── internal/
│   ├── auth/                # JWT issuance, refresh rotation, Argon2id hashing
│   ├── users/               # User account entity, profile retrieval, status updates
│   ├── roles/               # Role definitions (admin, learner) and checks
│   ├── invitations/         # Invitation token generation, verification, consumption
│   ├── courses/             # Learning path and course definitions
│   ├── modules/             # Course modules and sequence order
│   ├── lessons/             # Markdown lesson storage, prerequisites, publication
│   ├── notes/               # User personal notes, tagging, full-text search
│   ├── bookmarks/           # Content bookmarks (lessons/challenges)
│   ├── challenges/          # Coding challenge statements, metadata, test cases
│   ├── submissions/         # Submission records, history, queue stubs
│   ├── progress/            # Lesson progress, streaks, user activity logs
│   ├── admin/               # Administrative aggregations and metrics
│   ├── middleware/          # HTTP middlewares: Auth, RBAC, RateLimit, CORS, Slog
│   ├── config/              # Environment configuration loader with validation
│   ├── database/            # pgxpool setup, lifecycle, migrations runner
│   ├── platform/            # Shared utilities (crypto random, response envelopes, errors)
│   └── health/              # Liveness and readiness HTTP handlers
├── migrations/              # golang-migrate SQL migration files (up/down)
├── queries/                 # SQLC input queries (.sql)
├── sqlc.yaml                # SQLC configuration
├── go.mod
├── go.sum
└── Dockerfile
```

---

## 3. Frontend Architecture

### 3.1 Technology Stack & Decisions
- **Framework:** React 19 + TypeScript.
- **Bundler:** Vite with Fast Refresh and optimized production chunking.
- **Styling:** Tailwind CSS with CSS variables for dark-first developer theme tokens.
- **UI Primitives:** Radix UI primitives and shadcn/ui component patterns.
- **State & Data Fetching:** TanStack Query (React Query) for server state caching, background invalidation, and optimistic updates.
- **Form Handling:** React Hook Form coupled with Zod schema validation.
- **Routing:** React Router v6+ supporting nested layout trees and role-based route guards (`RequireAuth`, `RequireAdmin`).
- **Code & Markdown:** Monaco Editor for Go challenge coding; safe Markdown rendering using AST parsing with sanitization against XSS.

### 3.2 Frontend Directory Structure
```
frontend/src/
├── app/
│   ├── App.tsx              # Root component with Providers (QueryClient, Auth, Theme)
│   └── routes.tsx           # Route configuration and guards
├── components/
│   ├── ui/                  # Reusable UI primitives (Button, Input, Dialog, Card, etc.)
│   ├── layout/              # AppLayout, Sidebar, Navbar, AdminLayout, AuthLayout
│   └── shared/              # StatusBadge, EmptyState, ErrorBoundary, LoadingSpinner
├── features/
│   ├── auth/                # LoginForm, useAuth hook, auth service
│   ├── invitations/         # AcceptInvitationForm, admin invitation tables/modals
│   ├── dashboard/           # Learner dashboard cards, streak stats, continue learning
│   ├── learning-paths/      # Learning path listing and roadmaps
│   ├── courses/             # Course overview and module syllabus
│   ├── lessons/             # Lesson reader, markdown display, progress completion
│   ├── notes/               # Notes editor, markdown preview, tagging, autosave
│   ├── bookmarks/           # Bookmarks list and toggles
│   ├── challenges/          # Challenge statement, Monaco editor, run/submit controls
│   ├── progress/            # Analytics charts, history view
│   └── admin/               # Admin overview, user management, course/lesson authoring
├── hooks/                   # Custom reusable React hooks (useDebounce, useLocalStorage)
├── lib/                     # Axios/Fetch API client instance, query client setup, utils
├── routes/                  # Route page components mapping to URLs
├── services/                # API client functions typed with TypeScript interfaces
└── types/                   # Shared TypeScript models (User, Invitation, Lesson, etc.)
```

---

## 4. Cross-Cutting Concerns

### 4.1 Request Lifecycle & Middleware Pipeline
Every incoming HTTP request flows through the following pipeline:
```
Incoming HTTP Request
  │
  ▼
1. RealIP Middleware         (extracts client IP considering trusted reverse proxies)
  │
  ▼
2. RequestID Middleware      (generates/extracts X-Request-ID and injects into context)
  │
  ▼
3. SlogLogger Middleware     (logs method, path, request_id, status, duration)
  │
  ▼
4. Recoverer Middleware      (catches panics, returns 500 error envelope, logs stack)
  │
  ▼
5. CORS Middleware           (evaluates Origin against configured allowlist)
  │
  ▼
6. BodyLimit Middleware      (enforces MaxBytesReader e.g. 1MB to prevent DoS)
  │
  ▼
7. RateLimiter Middleware    (rate-limits sensitive routes like login and register)
  │
  ▼
8. AuthVerifier Middleware   (validates Bearer JWT; injects Claims/User into context)
  │
  ▼
9. RBAC Middleware           (enforces required role, e.g. RequireAdmin)
  │
  ▼
Handler Method               (executes domain service logic)
  │
  ▼
JSON Response Envelope       ({"data": ..., "meta": ...} or {"error": ...})
```

### 4.2 Error Handling & Uniform Envelopes
All responses follow consistent schemas:
- **Success:**
  ```json
  {
    "data": { ... },
    "meta": { "timestamp": "2026-11-04T16:00:00Z" }
  }
  ```
- **Error:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "The request payload failed validation.",
      "details": [
        { "field": "email", "message": "Email address must be valid" }
      ]
    }
  }
  ```
Application errors define semantic error codes (`UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `INTERNAL_ERROR`), which the transport layer automatically maps to standard HTTP status codes.

---

## 5. Security & Isolation Boundary for Code Execution

A fundamental architectural principle of Go Academy is: **Never execute learner code inside the main API process or on the host operating system.**

### 5.1 Phased Execution Architecture (Phase 5)
In Phase 1, the `/challenges/{id}/run` and `/challenges/{id}/submit` endpoints are not active for code execution. In Phase 5, the execution engine is built as an independent worker:
1. **Submission Ingestion:** The Go API records the submission in PostgreSQL with status `queued` and returns the submission ID to the client immediately.
2. **Worker Claiming:** A detached worker process periodically polls or claims pending jobs via `SELECT ... FOR UPDATE SKIP LOCKED`.
3. **Sandbox Execution:** The worker invokes the Docker engine to run an ephemeral container from a dedicated, minimal runner image (`go-academy-runner:latest`):
   - User: `10001` (unprivileged `runner` user).
   - Read-only root filesystem (`--read-only`), with bounded `tmpfs` mounted at `/tmp` (maximum 64MB).
   - Network isolation: `--network none` (no egress or ingress).
   - Security options: `cap-drop=ALL`, `no-new-privileges:true`.
   - Resource quotas: CPU quota `1.0`, memory limit `128m`, pids limit `30`.
   - Wall-clock timeout: 5 seconds execution limit enforced by both container timeout and worker context.
4. **Bounded Output & Teardown:** Worker captures stdout/stderr up to a maximum of 64KB, tears down and removes the container (`--rm`), evaluates test results, and writes the final status (`accepted`, `wrong_answer`, `timeout`, `runtime_error`, `compile_error`) to PostgreSQL.
5. **No Secret Leakage:** The runner container has no access to environment secrets, database connection strings, or host volumes. Hidden test cases are evaluated inside the container test runner; only aggregate test counts and public test feedback are returned to the client.
