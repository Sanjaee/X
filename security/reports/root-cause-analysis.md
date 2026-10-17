# LostMedia Root-Cause Analysis (RCA) Report

**Audit Date**: October 17, 2026  
**Focus**: Engineering, Architectural, and Code-Level Root Causes  

---

## 1. Deep Dive Root Causes by Domain

### A. Business Logic & Authorization Integrity (`LM-SEC-001`, `LM-SEC-002`)
- **Underlying Cause**: Development velocity shortcuts. Test helper handlers (`SimulatePayment`) designed for local test suites and sandbox validation were exposed as standard production HTTP handlers without environment barriers (`NODE_ENV != "production"` or `config.IsProduction()`).
- **Signature Verification Flaw**: In `HandlePlisioWebhook`, defensive programming was misapplied: the developer wrote `if payload.VerifyHash != ""` assuming that Plisio might not always send signatures in testing. In production, this created an inadvertent bypass where omitting the signature completely avoided validation.
- **Architectural Solution**:
  1. Enforce strict environment guards on test routes.
  2. Invert signature validation: require `VerifyHash` unconditionally when webhook secret/key is configured. If empty or invalid, reject with HTTP 401/422.

### B. Input Validation & Web Security (`LM-SEC-003`, `LM-SEC-007`, `LM-SEC-011`)
- **Underlying Cause**: Implicit trust of user-supplied URLs and generic file handling.
- **Stored XSS**: Profile website input was treated as an arbitrary string and rendered directly in JSX as `<a href={profileUser.website}>`. Because JavaScript allows `javascript:` pseudoprotocols in `href`, clicking the anchor tag triggered script execution.
- **Directory Indexing**: Using `http.FileServer` on a directory without an index file triggers automatic directory tree generation in the Go standard library.
- **Media Ingestion**: The switch-case in `media.go` used a permissive `default:` that classified all unknown extensions as generic images.
- **Architectural Solution**:
  1. Implement canonical URL protocol sanitization: enforce `http://` or `https://` whitelist; automatically normalize plain domains (e.g. `example.com` -> `https://example.com`) and reject `javascript:`, `data:`, `vbscript:`.
  2. Implement custom filesystem wrapper or route interceptor for `/uploads/` that returns 404 for directory requests.
  3. Enforce strict MIME sniffing and extension whitelisting on all uploaded media.

### C. Traffic Management & Edge Inconsistencies (`LM-SEC-005`, `LM-SEC-010`, `LM-SEC-013`)
- **Underlying Cause**: Multi-tier reverse proxy mismatch.
- **Rate Limit Collision**: `global-nginx` correctly restored client IPs from Cloudflare via `real_ip_header CF-Connecting-IP`. However, it forwarded requests to `x-proxy` over internal Docker networking. Because `x-proxy` did not configure `set_real_ip_from` for the Docker subnet (`172.16.0.0/12`), it evaluated all clients as having the IP of `global-nginx`.
- **HTTP Downgrade**: Port 80 on `global-nginx` was combined in the same block as port 443 instead of dedicated to a 301 HTTPS redirect.
- **Architectural Solution**:
  1. Add Docker internal subnet ranges to `set_real_ip_from` in `x-proxy`.
  2. Separate port 80 into an unconditional `return 301 https://$host$request_uri;` block.
  3. Add dedicated rate-limiting middleware for auth endpoints (`/auth/login`, `/auth/register`).

### D. Operational Resilience & Secrets Hygiene (`LM-SEC-004`, `LM-SEC-006`, `LM-SEC-008`, `LM-SEC-009`)
- **Underlying Cause**: Missing automation and deployment validations.
- **Hardcoded Secret**: In NextAuth configuration, an inline string was used as a fallback. If the `.env` variable was omitted, NextAuth operated with a predictable key.
- **Backup Absence**: Operational guides documented manual commands, but no automated host-level cron schedule was ever installed.
- **Architectural Solution**:
  1. Fail-fast validation on critical secrets at startup: if `NEXTAUTH_SECRET` is unset, fail immediately.
  2. Add `Secure: true` to all production authentication cookies.
  3. Deploy automated cron backup script with retention and validation checks.
  4. Restrict `/health/metrics` to internal access.
