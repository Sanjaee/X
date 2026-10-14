# Security Architecture & Hardening Specifications

**Status:** Phase 3 Approved  
**Standard:** OWASP Top 10 Compliance for Social Platforms

---

## 1. Authentication & Credential Protection

### Password Hashing
- Passwords are never stored in plaintext or weak hashes (MD5/SHA1/SHA256).
- **Algorithm**: `bcrypt` with cost factor 12 (or `Argon2id` with `time=2, memory=64MB, threads=4`).
- Password validation rules enforced server-side: Minimum 8 characters, maximum 128 characters, reject common dictionary lists.

### Session Management & Cookie Hardening
- Session tokens are generated as high-entropy cryptographically secure random bytes (32 bytes `crypto/rand`, base64 encoded).
- Stored hashed (SHA-256) in PostgreSQL `sessions` table to prevent credential exposure in database dumps.
- Transmitted exclusively via cookies with the following attributes:
  ```
  Set-Cookie: session_token=<token>; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=2592000
  ```
- Instant revocation support: When a user logs out or clicks "Log out all sessions", the session records are deleted or marked invalid in DB and Redis.

---

## 2. Server-Side Authorization & Multi-Tenancy

Every incoming API request passes through a centralized `AuthMiddleware`:
1. Extract session token from cookie or `Authorization: Bearer <token>` header.
2. Verify token against Redis cache; fallback to PostgreSQL `sessions` JOIN `users`.
3. Inject authenticated `UserContext` into Go `context.Context`.

### Resource Ownership Enforcement
- **Post Deletion**:
  ```sql
  DELETE FROM posts WHERE id = $1 AND author_id = $2;
  ```
  Users can only delete their own posts, unless their user role is `admin` or `moderator`.
- **Direct Messaging Privacy**:
  Message access requires active membership in `conversation_members`:
  ```sql
  SELECT 1 FROM conversation_members WHERE conversation_id = $1 AND user_id = $2;
  ```
  Requests from non-members are rejected with `403 Forbidden`.

---

## 3. Defense-in-Depth Mechanisms

### SQL Injection Prevention
- All database queries exclusively utilize parameterized queries (`$1, $2, ...` in PostgreSQL via `database/sql` / `sqlc` / `pgx`).
- Dynamic SQL string concatenation is strictly prohibited.

### Cross-Site Scripting (XSS) Prevention
- React automatically escapes string bindings in JSX.
- Rich text (mentions `@user` and hashtags `#topic`) is tokenized into abstract syntax trees rather than using raw `dangerouslySetInnerHTML`.
- HTTP response headers enforced by reverse proxy:
  ```http
  Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: blob:; font-src 'self'; connect-src 'self' wss: https:;
  X-Content-Type-Options: nosniff
  X-Frame-Options: SAMEORIGIN
  Referrer-Policy: strict-origin-when-cross-origin
  ```

### Media Upload Validation
- File upload checks:
  1. File size limit: Maximum 10MB for images, 100MB for video.
  2. Magic byte inspection: Validated using `http.DetectContentType` to ensure actual binary matches declared extension.
  3. Storage isolation: Uploaded files are renamed to UUIDs (`{uuid}.{ext}`) to eliminate path traversal vulnerabilities.
