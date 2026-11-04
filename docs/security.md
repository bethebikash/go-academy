# Go Academy --- Security Architecture & Policies

**Document Version:** 1.0  
**Status:** Security Baseline Specification  

---

## 1. Threat Model & Security Principles

Go Academy hosts private user data, learning content, and will eventually execute untrusted user-submitted Go code. The core security principles are:
1. **Never Run Learner Code in API Process:** Untrusted execution is completely separated into a sandboxed environment.
2. **Never Persist or Log Plaintext Secrets:** Passwords, refresh tokens, and invitation tokens are stored strictly as cryptographic one-way hashes.
3. **Defense in Depth:** Security controls are enforced at multiple layers: network/CORS, HTTP transport, authentication middleware, domain authorization, and database queries.
4. **Least Privilege:** Each user and service account possesses only the minimum necessary permissions.

---

## 2. Authentication & Credential Security

### 2.1 Password Hashing with Argon2id
All passwords are hashed using **Argon2id** (RFC 9106 recommended profile):
- **Memory Cost (`m`):** 64 MB (65,536 KiB)
- **Time Cost (`t`):** 3 iterations
- **Parallelism (`p`):** 2 threads
- **Salt Length:** 16 cryptographically random bytes generated via `crypto/rand`
- **Derived Key Length:** 32 bytes
- **Format:** Standard encoded modular crypt format (`$argon2id$v=19$m=65536,t=3,p=2$...`)

### 2.2 Short-Lived JWT Access Tokens
- **Lifespan:** 15 minutes.
- **Algorithm:** HMAC-SHA256 (`HS256`) with a cryptographically strong secret (minimum 32 bytes).
- **Claims:**
  ```json
  {
    "sub": "7f0932c0-...",
    "email": "learner@example.com",
    "role": "learner",
    "status": "active",
    "iat": 1730736000,
    "exp": 1730736900,
    "jti": "d63f4b82-..."
  }
  ```
- Access tokens are transmitted via the standard `Authorization: Bearer <token>` header and kept only in frontend memory (never `localStorage` or `sessionStorage`).

### 2.3 Rotating Refresh Tokens & Reuse Detection
- **Token Format:** 32 cryptographically random bytes (`crypto/rand`) encoded as URL-safe base64.
- **Storage:** Only the SHA-256 hash of the token is persisted in PostgreSQL (`refresh_sessions.token_hash`).
- **Cookie Security:** Refresh tokens are transported via `Set-Cookie` with:
  - `HttpOnly: true` (inaccessible to JavaScript).
  - `Secure: true` (only sent over HTTPS; disabled only in local development).
  - `SameSite: Lax` (protects against CSRF on cross-site requests).
  - `Path: /api/v1/auth` (restricted to authentication endpoints).
- **Rotation Mechanics:**
  1. On each `/auth/refresh` request, the provided token is verified by hash lookup.
  2. If the token is valid, it is marked as revoked (`revoked_at = NOW()`).
  3. A new refresh token is generated within the same `token_family_id` and issued alongside a new JWT.
- **Reuse Detection (Token Theft Defense):**
  If an already revoked refresh token is presented, the system treats this as evidence of token compromise. **All active tokens sharing the same `token_family_id` are revoked immediately**, terminating any compromised session.

---

## 3. Registration & Invitation Security

### 3.1 Invitation Lifecycle
1. **Creation:** An administrator generates an invitation for a specific email address.
2. **Cryptographic Token:** The backend generates a 32-byte cryptographically random token (`crypto/rand`).
3. **Storage:** The SHA-256 hash of the token is stored in `invitations.token_hash`. The raw token is returned once to the administrator and never logged.
4. **Acceptance Transaction:**
   - When the user registers with the raw token, the backend hashes it and executes an atomic transaction:
     ```sql
     SELECT id, email, expires_at, accepted_at, revoked_at
     FROM invitations
     WHERE token_hash = $1
     FOR UPDATE;
     ```
   - Checks:
     - `revoked_at IS NULL`
     - `accepted_at IS NULL`
     - `expires_at > NOW()`
     - `LOWER(users.email) == LOWER(invitations.email)`
   - On success, `accepted_at` is set to `NOW()`, the user is created, and the learner role is assigned.
   - Any concurrent attempt to reuse the token fails with `INVITATION_ALREADY_USED`.

### 3.2 Secure CLI Admin Bootstrap
- There is **no public admin registration endpoint**.
- The initial administrator account is provisioned strictly through a command-line utility:
  ```bash
  go run cmd/admin/main.go bootstrap --email="admin@example.com" --username="admin"
  ```
- The bootstrap command verifies that no administrator account exists before creating one. If an admin exists, it safely aborts.

---

## 4. Authorization & Access Control

### 4.1 Role-Based Access Control (RBAC)
- Roles (`admin`, `learner`) are verified in backend middleware:
  ```go
  func RequireRole(role string) func(http.Handler) http.Handler
  ```
- Admin endpoints under `/api/v1/admin/*` reject any non-admin request with `403 Forbidden`.
- Hiding frontend UI controls is never relied upon for authorization.

### 4.2 Object Ownership Verification
- Every request accessing user-scoped data (`notes`, `bookmarks`, `submissions`, `progress`) validates that `resource.user_id == current_user.id`.
- SQL queries filter directly by `user_id = $1` in the `WHERE` clause, preventing IDOR (Insecure Direct Object Reference) vulnerabilities.

---

## 5. Transport, Network & Injection Defenses

### 5.1 SQL Injection Defense
- 100% of database interactions are compiled via **SQLC** using typed queries and parameterized placeholders (`$1`, `$2`, etc.).
- No dynamic query concatenation is permitted in application code.

### 5.2 Cross-Site Scripting (XSS) Prevention
- Lesson content and personal notes are formatted in Markdown.
- Frontend rendering sanitizes all parsed Markdown AST output using **DOMPurify** or React's safe rendering tree.
- Raw HTML tags in user notes and lessons are escaped or stripped.

### 5.3 Cross-Site Request Forgery (CSRF) & CORS
- API applies strict CORS headers matching the configured frontend origin:
  - `Access-Control-Allow-Origin: https://academy.example.com`
  - `Access-Control-Allow-Credentials: true`
  - `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
- Refresh token cookies use `SameSite=Lax` and path restriction `/api/v1/auth`.

### 5.4 Rate Limiting & DoS Protection
- **Request Body Limits:** Handlers enforce `http.MaxBytesReader` (e.g. 1MB max body size) to prevent memory exhaustion attacks.
- **Endpoint Rate Limiting:**
  - Login attempts: max 5 requests per minute per IP.
  - Registration attempts: max 5 requests per minute per IP.
  - Refresh requests: max 30 requests per minute per IP.

---

## 6. Code Execution Sandbox Isolation (Phase 5 Roadmap)

The future Go execution engine executes untrusted code submitted by learners. It will strictly abide by the following isolation controls:

1. **Dedicated Host / Isolated Process:** Runs in a separate worker process on an isolated Docker host.
2. **Container Restrictions:**
   - **User:** Non-root (`uid 10001:gid 10001`).
   - **Filesystem:** Read-only root filesystem (`--read-only`), temporary writable `tmpfs` mounted at `/tmp` (limit 64MB, `noexec,nodev,nosuid`).
   - **Networking:** `--network none` (complete network isolation; no network device).
   - **Privileges:** `--cap-drop=ALL` (drops all Linux capabilities), `--security-opt=no-new-privileges:true`.
   - **Resource Quotas:**
     - CPU: `--cpus=1.0`
     - Memory: `--memory=128m --memory-swap=128m`
     - PIDs: `--pids-limit=30` (prevents fork bombs)
   - **Timeouts:** Container wall-clock timeout of 5.0 seconds enforced by execution runner context.
   - **Output Truncation:** Stdout/stderr buffers capped at 64KB to prevent storage exhaustion.
3. **No Socket Sharing:** The Docker daemon socket (`/var/run/docker.sock`) is never mounted inside the API container or runner container.
4. **Hidden Test Cases:** Hidden test inputs and expected outputs are never returned to the client; only pass/fail status and execution time are communicated.

---

## 7. Logging & Auditing

- **Structured Logging:** Go's standard `log/slog` library is used with JSON output.
- **Log Sanitization:** Sensitive attributes (passwords, JWTs, refresh tokens, invitation raw secrets, cookie headers) are explicitly scrubbed and never written to logs.
- **Audit Logging:** Administrative actions (invitation generation, invitation revocation, user status changes, content publication) emit dedicated structured audit events including actor ID, target ID, action, timestamp, and IP address.
