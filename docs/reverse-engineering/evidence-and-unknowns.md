# Reverse Engineering Evidence, Limitations, and Unknowns

**Status:** Phase 1 Complete  
**Standard:** Strict Empirical Integrity (No Invented Internals)

---

## 1. Verified Evidence Ledger

Every factual claim in this project is anchored to an observable artifact:

| Artifact / Evidence | SHA-256 / Location | Key Verified Fact |
| :--- | :--- | :--- |
| `https://x.com/` HTML | Dynamic DOM stream | Uses TanStack Router streaming SSR, Seroval streaming protocol, and Tailwind CSS. |
| `https://abs.twimg.com/x-web/x-web/assets/styles-BPdMqb99.css` | Production CSS bundle | Explicit Tailwind CSS v4.3.3 comment header, custom layout CSS properties (`--layout-width-primary: 600px`, `--layout-width-rail-labeled: 275px`, `--layout-width-right: 420px`). |
| `https://x.com/manifest.json` | Web App Manifest | Declared PWA shortcuts (`/compose/post`, `/explore`, `/notifications`, `/messages`), FCM sender ID, supported media types (`image/jpeg`, `image/png`, `image/gif`, `video/mp4`, `video/quicktime`). |
| `https://x.com/NASA` | Public User Profile | Profile layout with banner (3:1), circular avatar (`136px`), follower count format, subtabs (`/with_replies`, `/media`, `/reposts`). |
| `https://x.com/NASA/status/2110371804959191288` | Focal Post Status | View counter format (`41.2K Views`), full timestamp (`2:07 PM · Oct 14, 2026`), nested reply thread structure. |
| Media Drafting Lifecycle | X Web UI Composer | Observable client-side draft staging via object URLs before commit; deferred network egress until explicit Post user action. |
| Cloudflare R2 Bucket | `lostmedia` / S3 API | High-fidelity progressive image compression with bilinear downsampling and public CDN edge delivery via `pub-a638ff524f744502b575203fb48d39ae.r2.dev`. |
| Plisio Payment Gateway | `PLISIO_API_KEY` / IPN Webhooks | Cryptocurrency invoice generation and authoritative IPN webhook signature verification for subscription and verified badge activation. |
| REA CLI `rea-agents@6.5.0` | Local npm cache | Version 6.5.0 verified with 153 CLI commands and 207 MCP tools. |


---

## 2. Documented Unknowns

In compliance with reverse engineering methodology, the following elements are explicitly marked **UNKNOWN** because they cannot be determined from public external access:

1. **Backend Service Topology**:
   - The private internal microservice graph behind `api.x.com` cannot be observed.
   - Internal inter-service protocols (gRPC, Finagle, Kafka) cannot be directly inspected.
2. **Database Engine & Physical Schema**:
   - The internal database technology (whether PostgreSQL, MySQL, Cassandra, or proprietary distributed storage) is unobservable from HTTP API responses.
   - Exact internal SQL tables, column names, and partition strategies are unknown.
3. **Recommendation Algorithm Weights**:
   - The exact scoring formula, neural graph embeddings, and real-time candidate generation filters used for the "For you" feed are proprietary and unobservable.
   - *Implementation Strategy*: We will implement a clear, reproducible time-decay engagement ranking algorithm for our own "For you" feed, fully documented in our architecture specs.

---

## 3. Analysis Limitations

- **Rate Limits & Bot Protection**: X actively deploys Cloudflare and Arkose Labs challenge barriers on repeated automated queries. All investigations were performed via non-evasive passive inspection of public pages.
- **Client Minification**: Production JavaScript assets (`entry-client-logged-out-BJDlp_4g.js`) are aggressively minified and bundled without publicly accessible source maps. However, React component markers, hydration signatures, and CSS variable declarations remain clearly identifiable.
