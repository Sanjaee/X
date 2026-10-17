# Platform Owner Security, Bootstrap & Emergency Disaster Recovery Protocol

## 1. Overview & Security Architecture

The platform implements a strict separation between **Administrative Authorization Roles** and **Membership Benefit Tiers**:

- **Authorization Roles**: `member`, `mod`, `god`, `owner`
- **Membership Tiers**: `free`, `vip`, `mvp`

The `owner` role is the highest administrative authority in the system. It carries platform-wide governance privileges, including RBAC configuration, permission assignments, membership quota adjustments, user moderation, and audit trail inspection.

---

## 2. Owner Account Identification & Bootstrap Mechanism

### 2.1 Designated Primary Owner Candidate
- **Email**: `afrizaahmad18@gmail.com`

### 2.2 Security Invariants during Bootstrap
1. **Strict Authentication**: The account is **never** granted the `owner` role simply because an incoming request contains the email in a body, query parameter, or header.
2. **Provider Verification**: The user must complete an authentic authentication flow (e.g., verified Google OAuth 2.0 or verified credentials with `email_verified_at` confirmed).
3. **Email Normalization**: The email is always trimmed of whitespace and converted to lowercase (`strings.ToLower(strings.TrimSpace(email))`).
4. **Idempotent Provisioning**: During authentication (`Login`, `GoogleAuth`, `GetMe`), the backend checks if the authenticated user has the designated owner email. If so, `EnsureOwnerBootstrap` executes idempotently:
   - Queries or inserts the `owner` role in `roles`.
   - Assigns the `owner` role in `user_roles` without modifying or duplicating records if already present.
   - Assigns the highest tier (`mvp`) in `user_memberships`.
   - Records an immutable audit log entry in `audit_logs` with action `OWNER_BOOTSTRAP_EXECUTED`.
5. **No Public Endpoint**: There is no public or unauthenticated API endpoint that can assign, elevate, or claim the `owner` role.

---

## 3. Owner Immutability & Protection Rules

To prevent accidental platform lockout or malicious hostile takeover:

1. **Immunity to Role Deletion**:
   - The `owner` role has `is_system = TRUE`.
   - `DELETE /api/v1/admin/roles/{roleId}` explicitly returns `400 Bad Request` ("system roles cannot be deleted") if the `owner` role is targeted.
2. **Immunity to Role Demotion / Revocation via Admin API**:
   - `PATCH /api/v1/admin/users/{userId}/roles` strictly verifies if the target user is an Owner or if the `owner` role is being modified without authorization.
   - Standard administrators and moderators cannot revoke or replace an Owner's role.
3. **Immunity to Account Suspension**:
   - `PATCH /api/v1/admin/users/{userId}/status` rejects attempts to suspend or deactivate the primary Owner account.
4. **Self-Modification Block**:
   - Administrators cannot use administrative endpoints to elevate their own personal account or reassign their own permissions.

---

## 4. Emergency Disaster Recovery Procedure

If the primary Owner account credentials are lost, inaccessible, or the authentication provider encounters an outage, use the following operational procedures to restore access securely.

### Scenario A: OAuth Provider Credentials Lost / Domain Reassignment
If the Owner's email address needs to be transferred to a new verified corporate or administrative email:

1. **Access the Production Database Container**:
   ```bash
   # Connect to the VPS host
   ssh -i <identity_key> <user>@<host_ip>

   # Execute interactive psql inside the PostgreSQL container
   docker exec -it x-postgres psql -U media_user -d media_prod
   ```

2. **Verify Identity of the Replacement User**:
   Ensure the new user has already registered and verified their account via the platform login:
   ```sql
   SELECT id, email, username, email_verified_at, status 
   FROM users 
   WHERE email = 'new_owner@example.com';
   ```

3. **Transfer Owner Role & Record Audit Log**:
   Execute the recovery transaction:
   ```sql
   BEGIN;

   -- 1. Ensure Owner role ID
   SELECT id FROM roles WHERE key = 'owner';

   -- 2. Assign Owner role to new user
   INSERT INTO user_roles (user_id, role_id, assigned_by, reason)
   VALUES (
     (SELECT id FROM users WHERE email = 'new_owner@example.com'),
     (SELECT id FROM roles WHERE key = 'owner'),
     (SELECT id FROM users WHERE email = 'new_owner@example.com'),
     'EMERGENCY_DISASTER_RECOVERY_MANUAL_PROVISION'
   )
   ON CONFLICT (user_id, role_id) DO NOTHING;

   -- 3. Assign MVP membership
   INSERT INTO user_memberships (user_id, plan_id, status, starts_at, expires_at)
   VALUES (
     (SELECT id FROM users WHERE email = 'new_owner@example.com'),
     (SELECT id FROM membership_plans WHERE key = 'mvp'),
     'active',
     NOW(),
     NOW() + INTERVAL '10 years'
   )
   ON CONFLICT (user_id, plan_id) DO NOTHING;

   -- 4. Record audit trail
   INSERT INTO audit_logs (actor_user_id, action, target_type, target_id, reason)
   VALUES (
     (SELECT id FROM users WHERE email = 'new_owner@example.com'),
     'EMERGENCY_OWNER_RECOVERY',
     'user',
     (SELECT id FROM users WHERE email = 'new_owner@example.com')::text,
     'Manual root CLI disaster recovery transfer'
   );

   COMMIT;
   ```

4. **Verify Access**:
   Log in with `new_owner@example.com`. The user will receive full administrative access to `/admin` and all owner-guarded resources.

---

### Scenario B: Database Corrupted or Restored from Backup
When the database is restored from cold backup:

1. Run the database migration script:
   ```bash
   docker exec -it x-api /app/server -migrate
   # Or execute the migration SQL directly:
   docker exec -i x-postgres psql -U media_user -d media_prod < services/api/migrations/000006_rbac_roles_memberships_quotas.up.sql
   ```
2. The migration script idempotently creates all system roles (`member`, `vip`, `mvp`, `mod`, `god`, `owner`), maps all 26 default permissions, and bootstraps `afrizaahmad18@gmail.com` if existing.
3. Upon first login with `afrizaahmad18@gmail.com`, the API server automatically ensures role assignment and logs the bootstrap action.

---

## 5. Audit Trail & Compliance

Every action taken by an Owner or Moderator is captured in the `audit_logs` table:

| Field | Description |
|---|---|
| `id` | Unique UUID of the audit log entry |
| `actor_user_id` | UUID of the user who performed the administrative action |
| `action` | Action descriptor (e.g. `OWNER_BOOTSTRAP_EXECUTED`, `ROLE_UPDATED`, `POST_MODERATED_DELETE`) |
| `target_type` | Entity type affected (`role`, `permission`, `user`, `plan`, `post`) |
| `target_id` | Identifier of the affected entity |
| `reason` | Administrative justification provided for the action |
| `created_at` | UTC timestamp of the action |

Audit logs are viewable directly from the Owner Dashboard at `/admin` under the **Audit Trail** tab.
