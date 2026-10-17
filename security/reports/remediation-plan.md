# LostMedia Security Remediation Plan

**Audit Date**: October 17, 2026  
**Implementation Strategy**: Zero-Downtime Hotfix & Deployment  

---

## 1. Remediation Roadmap by Priority

```
[ P0 Critical ]
├── LM-SEC-001: Gate /subscriptions/simulate-payment behind non-production check & user ownership
└── LM-SEC-002: Mandate cryptographic verify_hash in HandlePlisioWebhook

[ P1 High ]
├── LM-SEC-004: Eliminate hardcoded NextAuth secret fallback; fail fast if unset
├── LM-SEC-005: Add Docker subnet real_ip restoration in x-proxy to fix rate limiting
├── LM-SEC-007: Enforce URL protocol whitelisting (http/https) in UpdateProfile & frontend JSX
└── LM-SEC-008: Deploy automated daily database backup script to VPS crontab

[ P2 Medium ]
├── LM-SEC-003: Suppress public directory indexing on /uploads/ in Go file server
├── LM-SEC-006: Set Secure: true on session authentication cookies
├── LM-SEC-011: Whitelist accepted media extensions (.jpg, .jpeg, .png, .webp, .gif, .mp4, .mov, .webm)
└── LM-SEC-013: Add dedicated rate limiter (10 req/min) for login & register

[ P3 Low / Informational ]
├── LM-SEC-009: Restrict /health/metrics to localhost / internal networks
├── LM-SEC-010: Redirect HTTP port 80 to HTTPS 301, hide X-Powered-By, add CSP
└── LM-SEC-012: Document container root migration plan for future build cycles
```

---

## 2. File Change Specifications

### A. `services/api/internal/handler/handler.go`
- **LM-SEC-001**: In `SimulatePayment`, add check for `os.Getenv("ALLOW_PAYMENT_SIMULATION") == "true"` or `os.Getenv("NODE_ENV") != "production"`. Return 403 Forbidden otherwise. Also check `order.UserID != user.ID`.
- **LM-SEC-002**: In `HandlePlisioWebhook`, if `h.plisioSvc != nil`, require `payload.VerifyHash != ""`. If empty, return 401 Unauthorized.
- **LM-SEC-006**: In `setSessionCookie`, add `Secure: true`.
- **LM-SEC-007**: In `UpdateProfile`, validate and sanitize `req.Website` to require `http://` or `https://` prefix, rejecting `javascript:` or `data:`.

### B. `services/api/cmd/server/main.go`
- **LM-SEC-003**: Replace raw `http.FileServer` with a guarded handler that returns 404 for directory requests.
- **LM-SEC-009**: Guard `/health/metrics` so it requires an internal header or localhost origin.
- **LM-SEC-013**: Apply dedicated rate limiting middleware for `/auth/login` and `/auth/register`.

### C. `services/api/internal/media/media.go`
- **LM-SEC-011**: Add strict extension whitelist and MIME type validation in `ProcessAndUpload` and `saveLocalFallback`.

### D. `apps/web/src/app/[username]/page.tsx`
- **LM-SEC-007**: Sanitize website rendering in JSX: ensure URL scheme starts with `http://` or `https://`, prepend `https://` if missing, reject `javascript:`, and set `rel="noopener noreferrer"`.

### E. `apps/web/src/lib/auth.ts`
- **LM-SEC-004**: Replace fallback with fail-fast assertion:
  ```ts
  const secret = process.env.NEXTAUTH_SECRET;
  if (!secret && process.env.NODE_ENV === "production") {
    throw new Error("FATAL: NEXTAUTH_SECRET is required in production");
  }
  ```

### F. `deploy/nginx/nginx.conf` (`x-proxy`)
- **LM-SEC-005**: Add:
  ```nginx
  set_real_ip_from 10.0.0.0/8;
  set_real_ip_from 172.16.0.0/12;
  set_real_ip_from 192.168.0.0/16;
  real_ip_header X-Forwarded-For;
  real_ip_recursive on;
  ```

### G. `deploy/nginx/global-nginx.conf` (`global-nginx`)
- **LM-SEC-010**: Separate port 80 to redirect to `https://$host$request_uri;`. Add `proxy_hide_header X-Powered-By;` and security headers.

### H. `scripts/backup-db.sh`
- **LM-SEC-008**: Automated daily backup script for `x_prod`, compressed with gzip, checksummed with SHA256, and managed with 14-day retention.
