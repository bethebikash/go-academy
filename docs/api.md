# Go Academy --- REST API Specification

**Document Version:** 1.0  
**Base URL:** `/api/v1`  
**Protocol:** HTTPS / JSON REST  

---

## 1. General API Conventions

### 1.1 Request and Response Format
- All request and response bodies use JSON (`Content-Type: application/json`).
- Endpoints return standard HTTP status codes (`200 OK`, `201 Created`, `204 No Content`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `409 Conflict`, `422 Unprocessable Entity`, `429 Too Many Requests`, `500 Internal Server Error`).

### 1.2 Response Envelopes
Every API response is formatted using a standardized envelope.

#### Success Envelope
```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2026-11-04T16:00:00Z",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total_items": 150,
      "total_pages": 8
    }
  }
}
```

#### Error Envelope
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "The request payload failed validation.",
    "details": [
      {
        "field": "email",
        "message": "A valid email address is required."
      }
    ]
  }
}
```

Standard error codes:
- `UNAUTHENTICATED`: Missing, invalid, or expired authentication token.
- `FORBIDDEN`: The user does not have permission to access the resource.
- `VALIDATION_ERROR`: Field validation failed.
- `NOT_FOUND`: Target resource does not exist.
- `CONFLICT`: Unique constraint violation (e.g. email or username already taken).
- `INVITATION_EXPIRED`: The provided invitation token has expired.
- `INVITATION_REVOKED`: The invitation has been revoked.
- `INVITATION_ALREADY_USED`: The invitation has already been accepted.
- `ACCOUNT_DISABLED`: The account has been deactivated.
- `RATE_LIMIT_EXCEEDED`: Too many requests submitted.
- `INTERNAL_ERROR`: Unexpected server error.

---

## 2. Authentication & Session Flow

- **Access Token:** Short-lived JWT (15 minutes expiry) passed in the `Authorization` header:
  ```http
  Authorization: Bearer <jwt_access_token>
  ```
- **Refresh Token:** High-entropy opaque cryptographically random string (32 bytes) stored inside a secure `HttpOnly`, `SameSite=Lax`, `Path=/api/v1/auth`, `Secure` cookie named `refresh_token`.
- **Token Rotation:** Each refresh request burns the old refresh token and issues a new pair (JWT + new rotating refresh token).
- **Reuse Detection:** If an already-burned refresh token is presented, the entire token family is revoked immediately.

---

## 3. Endpoints Contract

### 3.1 Authentication (`/auth`)

#### `POST /auth/login`
Authenticates a user with email and password.
- **Request:**
  ```json
  {
    "email": "learner@example.com",
    "password": "CorrectHorseBatteryStaple1!"
  }
  ```
- **Response (200 OK):**
  - Sets `Set-Cookie: refresh_token=<opaque_token>; HttpOnly; Secure; SameSite=Lax; Path=/api/v1/auth; Max-Age=604800`
  - Body:
    ```json
    {
      "data": {
        "access_token": "eyJhbGciOi...",
        "expires_in": 900,
        "user": {
          "id": "7f0932c0-...",
          "email": "learner@example.com",
          "username": "learner1",
          "role": "learner",
          "status": "active"
        }
      }
    }
    ```

#### `POST /auth/register`
Accepts an invitation and registers a new account in a single atomic transaction.
- **Request:**
  ```json
  {
    "invitation_token": "8f9d02a...",
    "username": "gopher123",
    "password": "SecurePassword123!"
  }
  ```
- **Response (201 Created):**
  - Sets `Set-Cookie: refresh_token=...`
  - Body:
    ```json
    {
      "data": {
        "access_token": "eyJhbGciOi...",
        "expires_in": 900,
        "user": {
          "id": "c1f7b820-...",
          "email": "friend@example.com",
          "username": "gopher123",
          "role": "learner",
          "status": "active"
        }
      }
    }
    ```

#### `POST /auth/refresh`
Rotates refresh token and returns a new access token.
- **Headers/Cookies:** Requires `refresh_token` cookie.
- **Response (200 OK):**
  - Updates `Set-Cookie: refresh_token=<new_opaque_token>; ...`
  - Body:
    ```json
    {
      "data": {
        "access_token": "eyJhbGciOi...",
        "expires_in": 900
      }
    }
    ```

#### `POST /auth/logout`
Revokes the current refresh session and clears the cookie.
- **Response (204 No Content):**
  - Sets `Set-Cookie: refresh_token=; Max-Age=0; ...`

#### `GET /auth/me`
Fetches the currently authenticated user profile.
- **Headers:** `Authorization: Bearer <jwt_access_token>`
- **Response (200 OK):**
  ```json
  {
    "data": {
      "id": "7f0932c0-...",
      "email": "learner@example.com",
      "username": "learner1",
      "role": "learner",
      "status": "active",
      "timezone": "UTC",
      "created_at": "2026-11-04T12:00:00Z"
    }
  }
  ```

---

### 3.2 Invitations (Admin Only: `/admin/invitations`)

#### `POST /admin/invitations`
Generates a new invitation.
- **Headers:** `Authorization: Bearer <admin_token>`
- **Request:**
  ```json
  {
    "email": "friend@example.com",
    "expires_in_days": 7
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "data": {
      "id": "e3a890b2-...",
      "email": "friend@example.com",
      "invitation_token": "raw_token_presented_only_once_to_admin",
      "invitation_url": "https://academy.example.com/register?token=raw_token_presented_only_once_to_admin",
      "expires_at": "2026-11-11T16:00:00Z",
      "created_at": "2026-11-04T16:00:00Z"
    }
  }
  ```

#### `GET /admin/invitations`
Lists invitations with pagination and status filter.
- **Query Params:** `?page=1&limit=20&status=pending` (pending, accepted, expired, revoked)
- **Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "e3a890b2-...",
        "email": "friend@example.com",
        "status": "pending",
        "expires_at": "2026-11-11T16:00:00Z",
        "accepted_at": null,
        "revoked_at": null,
        "invited_by": "018f23...",
        "created_at": "2026-11-04T16:00:00Z"
      }
    ],
    "meta": {
      "pagination": { "page": 1, "limit": 20, "total_items": 1 }
    }
  }
  ```

#### `DELETE /admin/invitations/{id}`
Revokes an existing pending invitation.
- **Response (200 OK):**
  ```json
  {
    "data": {
      "id": "e3a890b2-...",
      "status": "revoked",
      "revoked_at": "2026-11-04T16:05:00Z"
    }
  }
  ```

---

### 3.3 Users (Admin Only: `/admin/users`)

#### `GET /admin/users`
Lists users with pagination and role/status filtering.
- **Query Params:** `?page=1&limit=20&search=john`
- **Response (200 OK):** Returns list of user records.

#### `PATCH /admin/users/{id}/status`
Activates or deactivates a user account.
- **Request:**
  ```json
  {
    "status": "disabled"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "data": {
      "id": "7f0932c0-...",
      "status": "disabled",
      "updated_at": "2026-11-04T16:10:00Z"
    }
  }
  ```

---

### 3.4 Learning Paths, Courses & Lessons

#### `GET /learning-paths`
Lists published learning paths for learners.
- **Response (200 OK):** Array of learning path summaries.

#### `GET /courses/{slug}`
Retrieves a course syllabus with modules and lessons.
- **Response (200 OK):** Course hierarchy with module and lesson titles, summaries, and prerequisites.

#### `GET /lessons/{slug}`
Retrieves lesson markdown content and user's completion status.
- **Response (200 OK):** Full lesson content, prerequisites, and user progress (`started`, `completed`).

#### `PUT /lessons/{id}/progress`
Idempotently marks a lesson as completed.
- **Request:** `{"status": "completed"}`
- **Response (200 OK):** Updated progress status and recorded timestamp.

---

### 3.5 Personal Notes & Bookmarks

#### `GET /notes`
Lists private notes belonging to authenticated user.
- **Query Params:** `?tag=concurrency&lesson_id=...&pinned=true`
- **Response (200 OK):** Array of user notes.

#### `POST /notes`
Creates a new personal note.
- **Request:**
  ```json
  {
    "lesson_id": "8f39...",
    "title": "Notes on Channels",
    "content": "Buffered channels block on write when full.",
    "pinned": true
  }
  ```
- **Response (201 Created):** Created note record.

#### `PATCH /notes/{id}`
Updates a personal note with optimistic concurrency check.
- **Request:**
  ```json
  {
    "title": "Updated Title",
    "content": "Updated content...",
    "version": 2
  }
  ```
- **Response (200 OK):** Updated note with incremented version.

#### `POST /bookmarks`
Creates a bookmark for a lesson or challenge.
- **Request:**
  ```json
  { "lesson_id": "8f39..." }
  ```
- **Response (201 Created):** Created bookmark record.

---

### 3.6 Challenges & Submissions (Phase 4 & 5 Preview)

#### `GET /challenges`
Lists published challenges with difficulty and topic filters.

#### `GET /challenges/{slug}`
Returns challenge statement, constraints, examples, and starter code (hidden test cases are omitted).

#### `POST /challenges/{id}/run` *(Future - Phase 5)*
Runs sample test cases against the sandbox runner.

#### `POST /challenges/{id}/submit` *(Future - Phase 5)*
Submits code for complete grading against all test cases.

---

### 3.7 Health Endpoints

#### `GET /health/live`
Kubernetes/Docker liveness probe.
- **Response (200 OK):** `{"status": "ok"}`

#### `GET /health/ready`
Readiness probe verifying database connectivity.
- **Response (200 OK):** `{"status": "ready", "database": "up"}`
- **Response (503 Service Unavailable):** `{"status": "unready", "database": "down"}`
