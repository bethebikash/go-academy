# Go Academy --- Product & Technical Specification

**Document type:** Product Requirements + Software Architecture + Implementation Guide  
**Version:** 1.0  
**Date:** 4 November 2026  
**Product:** Go Academy  
**Status:** Approved product direction; implementation to proceed incrementally  

---

## 1. Executive summary

Go Academy is a multi-user, interactive learning platform focused on teaching the Go programming language. It combines structured learning paths, detailed lessons, personal notes, coding challenges, secure code execution, explanations, and learning analytics.

The product begins as a platform for its owner and will later be shared with invited friends. It must therefore support multiple independent user accounts, authorization, private user data, and administrator-managed learning content from the first release.

The application should be built as a **modular monolith**, not as microservices. The codebase should be straightforward to run locally, test, maintain, and extend.

### Approved product decisions

| Decision | Selection |
| --- | --- |
| Product name | Go Academy |
| Backend | Go |
| Frontend | React + TypeScript |
| Database | PostgreSQL |
| API style | Versioned REST API |
| Registration | Invite-only |
| Roles | Admin and Learner |
| Content management | Admin-managed through the application |
| Code execution | Separate, isolated Docker-based runner |
| UI direction | Dark-first developer-focused interface |
| Architecture | Modular monolith |
| Repository | Monorepo |
| Initial delivery | Incremental phases, with Phase 1 as the first implementation target |

---

## 2. Product vision and principles

### Vision
Provide a focused environment where learners can understand Go concepts, practice them through executable challenges, write and organize personal notes, and see their progress over time.

### Product principles
1. **Understand, then memorize:** lessons explain why a concept exists, not only its syntax.
2. **Practice is central:** each topic should connect to relevant exercises and challenges.
3. **Explain outcomes:** submissions should provide useful compiler feedback, test results, hints, and solution explanations.
4. **Personal learning space:** notes, bookmarks, submissions, and progress belong to individual users.
5. **Admin-managed curriculum:** authorized administrators create, edit, organize, and publish learning content.
6. **Secure by design:** never run learner code inside the main API process.
7. **Incremental delivery:** build and verify one milestone at a time.
8. **Avoid premature complexity:** begin with a modular monolith and introduce separate services only where isolation or scale requires them.

---

## 3. Users, roles, and access

### 3.1 Learner
A learner can:
- Sign in and sign out.
- View published learning paths, courses, modules, and lessons.
- Mark lessons complete and revisit them.
- Create, edit, delete, search, and organize private notes.
- Bookmark lessons and challenges.
- Attempt coding challenges.
- Submit Go solutions and view execution results.
- View hints and solution explanations when permitted by challenge rules.
- View personal progress, streaks, and learning history.
- Manage permitted profile and preference settings.

A learner cannot:
- Create or publish courses, lessons, or challenges.
- Manage users or invitations.
- Access another learner's private notes, submissions, or progress.
- Change their own role or account status.

### 3.2 Administrator
An administrator has learner capabilities and can additionally:
- Create, edit, archive, publish, and unpublish learning paths/courses.
- Manage modules and lessons.
- Manage challenge statements, hints, explanations, difficulty, and test cases.
- Preview content before publication.
- Create, list, revoke, and inspect invitations.
- List users and activate/deactivate accounts.
- View platform-level content and usage summaries.
- Manage supported taxonomy such as tags and categories.

Admin permissions must be enforced by the backend. Hiding an admin control in the frontend is not authorization.

### 3.3 Registration policy
- Public self-registration is disabled.
- An administrator creates an invitation for a specific email address.
- The invitee registers using the invitation link.
- The invitation is single-use and expires after a configurable period (default: seven days).
- The registered email must match the invited email.
- The first administrator is created using a secure CLI bootstrap command, never a public HTTP endpoint.
- No default admin password or hardcoded production credential is allowed.

---

## 4. Technology stack

Use maintained, compatible versions and pin dependencies appropriately.

### Backend
- Go
- Chi HTTP router
- PostgreSQL
- pgx
- SQLC for type-safe query generation
- golang-migrate for database migrations
- Argon2id for password hashing
- Short-lived JWT access tokens
- Rotating opaque refresh tokens
- Structured logging using Go's `log/slog` or an equivalent maintained logger
- OpenAPI documentation for the REST API

### Frontend
- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- React Router
- TanStack Query
- React Hook Form
- Zod
- Monaco Editor for challenge coding
- A Markdown editor and safe Markdown renderer
- Syntax highlighting for Go code examples

### Infrastructure and quality
- Docker and Docker Compose
- PostgreSQL persistent volume for local development
- Environment-based configuration
- Git
- Go testing and `httptest`
- Vitest and React Testing Library
- Playwright for critical end-to-end flows
- CI workflow for formatting, linting, tests, and builds

Do not add Redis, Kubernetes, microservices, or a message broker unless a demonstrated requirement justifies them. A database-backed job queue or a simple worker can be considered for the first runner implementation.

---

## 5. High-level architecture

Use a modular monolith for the main product.

```text
Browser
  |
  v
React + TypeScript application
  |
  | HTTPS / JSON REST API
  v
Go API (modular monolith)
  |
  +-- Authentication and sessions
  +-- Users and roles
  +-- Invitations
  +-- Courses and learning paths
  +-- Modules and lessons
  +-- Notes and bookmarks
  +-- Challenges
  +-- Submissions and progress
  +-- Admin
  |
  v
PostgreSQL

Separate execution boundary (later phase)
Go API -> execution job -> isolated runner worker
                              |
                              v
                      restricted Docker container
                              |
                              v
                       result persisted in DB
```

### Architectural rules
- HTTP handlers parse requests, call application/service logic, and map results to HTTP responses.
- Business rules do not live inside handlers.
- Database access is implemented through generated SQLC queries and repository boundaries where useful.
- Modules should avoid importing each other's internal implementation details.
- Cross-module communication should use explicit interfaces or application services.
- Shared infrastructure code must remain small and generic.
- The code runner must be isolated from the API process and database credentials.
- Do not create microservices for ordinary CRUD modules.

---

## 6. Monorepo structure

```text
go-academy/
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go
│   │   └── admin/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/
│   │   ├── users/
│   │   ├── roles/
│   │   ├── invitations/
│   │   ├── courses/
│   │   ├── modules/
│   │   ├── lessons/
│   │   ├── notes/
│   │   ├── bookmarks/
│   │   ├── challenges/
│   │   ├── submissions/
│   │   ├── progress/
│   │   ├── admin/
│   │   ├── middleware/
│   │   ├── config/
│   │   ├── database/
│   │   ├── platform/
│   │   └── health/
│   ├── migrations/
│   ├── queries/
│   ├── sqlc.yaml
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   ├── layout/
│   │   │   └── shared/
│   │   ├── features/
│   │   │   ├── auth/
│   │   │   ├── invitations/
│   │   │   ├── dashboard/
│   │   │   ├── learning-paths/
│   │   │   ├── courses/
│   │   │   ├── lessons/
│   │   │   ├── notes/
│   │   │   ├── bookmarks/
│   │   │   ├── challenges/
│   │   │   ├── progress/
│   │   │   └── admin/
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── routes/
│   │   ├── services/
│   │   └── types/
│   ├── package.json
│   └── Dockerfile
├── docs/
│   ├── product-specification.md
│   ├── architecture.md
│   ├── database.md
│   ├── api.md
│   ├── security.md
│   ├── development.md
│   ├── content-authoring.md
│   └── deployment.md
├── docker-compose.yml
├── .env.example
├── .gitignore
└── README.md
```

---

## 7. Functional requirements

See implementation details in respective technical design documents.
