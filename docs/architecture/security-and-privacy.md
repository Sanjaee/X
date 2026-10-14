# Architecture: Security, Privacy, and Content Protection

## 1. Authorization Invariants
- **Identity Integrity:** The backend strictly derives user identity from validated sessions/JWT tokens in `r.Context()`. The client never passes or overrides `user_id`.
- **Private Account Visibility:** If `is_private = TRUE`, posts and followers are visible ONLY if `viewer_id == author_id` OR `follows(follower: viewer, following: author)` exists.
- **Block Protection:**
  - Blocked users cannot query the blocker's profile or posts (HTTP 403 Forbidden).
  - Blocked users cannot send messages to existing conversations with the blocker.
  - Adding a block dissolves mutual following immediately.
- **Direct Message Membership:** Any attempt to read or send messages in a `conversation_id` where the authenticated user is NOT an active member returns `403 Forbidden`.
- **Input Sanitization:** All post content and message bodies undergo HTML sanitization to neutralize stored XSS vectors while preserving legitimate links and mentions.
