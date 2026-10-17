# LostMedia Residual Risks & Architectural Limitations

**Audit Date**: October 17, 2026  
**Auditor**: REA Security & Resiliency Team  

---

## 1. Single-Node VPS Hardware Redundancy (Architectural Limitation)

- **Risk Description**: The entire platform currently resides on a single physical 4GB VPS (`103.89.2.68`).
- **Impact**: While container restart policies (`unless-stopped`), bounded memory limits, and automated backups provide high resilience against software and application errors, physical hardware faults, provider maintenance, or hypervisor failures will result in temporary unavailability until reboot.
- **Recommended Evolution**:
  1. Multi-AZ database replication (e.g., managed PostgreSQL with streaming read-replicas).
  2. Stateless container replication behind a distributed load balancer (HAProxy / AWS ALB / Cloudflare Load Balancing).

---

## 2. In-Memory Session Storage vs. Distributed JWT Revocation

- **Risk Description**: Sessions are stored in the PostgreSQL database and cached in memory. Revocation is instantaneous via database lookup.
- **Residual Risk**: High concurrent write volume on session logins creates IOPS on PostgreSQL.
- **Mitigation in Place**: SHA-256 token hashing with unique index lookup `sessions_token_hash_key`.
- **Future Enhancement**: Offload active session validation to Redis with automatic TTL expiry to reduce PostgreSQL read overhead.

---

## 3. Container Root User Execution (LM-SEC-012)

- **Risk Description**: While containers are network-isolated and do not share host mounts with sensitive paths, Docker processes run as UID 0 within the container namespace.
- **Status**: Documented with migration plan to non-root users (`appuser:1001`) in upcoming scheduled maintenance windows.
