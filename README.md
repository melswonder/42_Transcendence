*This project has been created as part of the 42 curriculum by hirwatan, sguruge, atashiro, ttanaka, kanahash.*

# Transcendence

A real-time online **Quoridor** platform — race your pawn across a 9×9 board while
placing walls to slow your opponent down. Built as the final project of the 42
Common Core (ft_transcendence).

## Team Information

| Member   | Role(s)                     | Responsibilities                                                                                                                     |
| -------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| hirwatan | Product Owner / Developer   | Product vision and backlog, validation of completed work; authentication (email+password, OAuth 2.0), Gin migration, HTTPS/Caddy, repository administration |
| ttanaka  | Project Manager / Developer | Planning, task tracking and coordination; Quoridor rules engine, real-time synchronization, remote play and reconnection logic       |
| sguruge  | Tech Lead / Developer       | Architecture (Clean Architecture layering) and code review; user management, friends & presence, statistics, achievements, spectator mode |
| atashiro | Developer                   | Public API with API keys and rate limiting, analytics dashboard filters and CSV/PDF exports                                          |
| kanahash | Developer                   | ORM data layer, internationalization (ja/en/fr), shared UI system (modals, auth screens), documentation                              |

## Description

Transcendence is a browser-based real-time multiplayer board game platform.
Two players face each other on a Quoridor board; every move is validated and
applied on the server, and the resulting state is broadcast to both players and
any spectators.

### Key Features

- **Server-authoritative Quoridor** — moves, jumps, wall placement and BFS path
  validation all run on the server; clients only render
- **Real-time Multiplayer** via WebSocket with reconnection grace, optimistic
  versioning (`gameVersion`) and idempotent retries (`actionId`)
- **Quick Matchmaking** with Elo rating, XP/levels and 12 achievements
- **Spectator Mode** — watch any live match mid-game with live spectator counts
- **User Authentication** with email/password (bcrypt hashed + salted) and Google OAuth 2.0
- **Player Profiles** with validated avatar uploads, friends and online status
- **Statistics Dashboard** with interactive charts, filters (date/mode/result/opponent) and CSV/PDF export
- **Match History & Leaderboard** with per-match rating changes
- **Public REST API** secured by hashed API keys with scopes, rate limiting and OpenAPI docs
- **Internationalization** — complete Japanese / English / French translations
- **HTTPS/WSS** for every browser connection (Caddy reverse proxy)
- **Privacy Policy & Terms of Service** pages in all three languages

## Instructions

### Prerequisites

- [Docker](https://www.docker.com/) (with Compose v2) and `make`
- Git
- A Google OAuth client — only for Google sign-in; email/password works without it
- No need to install Go or Node locally — everything runs inside containers

### Setup & Running

1. Clone the repository:

    ```bash
    git clone <repository-url>
    cd 42_Transcendence
    ```

2. Create a `.env` file from the example:

    ```bash
    cp .env.example .env
    ```

3. Configure the `.env` file with your settings. Only the Google OAuth
   credentials are required — every other variable falls back to the default
   in `docker-compose.yml`:

    ```env
    GOOGLE_CLIENT_ID=<your-client-id>.apps.googleusercontent.com
    GOOGLE_CLIENT_SECRET=<your-client-secret>
    ```

    For Google sign-in, register the authorized redirect URI in the
    [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
    (exact match required):

    ```
    https://localhost:8443/auth/google/callback
    ```

4. Build and start all services with a single command:

    ```bash
    make
    ```

5. Open **https://localhost** and **https://localhost:8443** once each and accept
   the locally-generated certificate (Caddy issues a self-signed local CA), then
   sign up and play.

### Services

| Service            | Port       | Description                                       |
| ------------------ | ---------- | ------------------------------------------------- |
| Caddy              | 443 / 8443 | HTTPS entrypoints: frontend (443), API/WSS (8443) |
| Frontend (Next.js) | 3000       | Web application UI (direct dev access)            |
| Backend (Go + Gin) | 4000       | REST API + WebSocket + SSE (direct dev access)    |
| PostgreSQL         | 5432       | Database                                          |
| atlas-dev          | 5433       | Scratch database for migration diffing only       |

Database migrations (Atlas) are applied automatically when the backend container
starts.

### Stopping

```bash
make down     # graceful shutdown (keeps data)
make fclean   # also remove database volumes
```

### Rebuilding after changes

Source directories are bind-mounted, so Go (air) and Next.js hot-reload
automatically. After changing dependencies or Dockerfiles:

```bash
make re
# frontend dependency changes also need the anonymous node_modules volume renewed:
docker compose build frontend && docker compose up -d -V frontend
```

## Technical Stack

| Layer            | Technology                        | Justification                                                                                     |
| ---------------- | --------------------------------- | ------------------------------------------------------------------------------------------------- |
| Frontend         | **Next.js 16** (React 19)         | App Router file-based routing; Server/Client component split keeps auth and data on the server    |
| UI               | **Mantine 9** + **Tailwind CSS 4**| Accessible ready-made components plus utility styling, coexisting through CSS cascade layers      |
| i18n             | **next-intl**                     | Cookie-based locale without URL changes; ICU messages for plurals and localized dates/numbers     |
| Backend          | **Go 1.26** + **Gin**             | Goroutines handle many concurrent WebSocket connections; Gin provides the route tree and middleware chain |
| Architecture     | **Clean Architecture**            | Game rules live in `domain` with zero HTTP/DB imports — the whole rules engine is unit-testable   |
| Realtime         | **coder/websocket** + SSE         | Server-authoritative state pushed as full snapshots; SSE refreshes statistics on match end        |
| ORM / Migrations | **GORM** + **Atlas**              | Struct-driven schema with versioned, reviewable SQL migrations; transactions guard multi-row updates |
| Database         | **PostgreSQL 16**                 | CHECK constraints, partial unique indexes and `citext` encode invariants at the database level    |
| Auth             | Cookie sessions + **bcrypt** + **OAuth 2.0 / OIDC** | Hashed+salted passwords; opaque session tokens stored only as SHA-256 hashes    |
| HTTPS            | **Caddy**                         | One small container terminates TLS (including WSS) with a local CA                                |
| Infra            | **Docker Compose**                | Single-command deployment for evaluation                                                          |

## Database Schema

### Tables

```
users
  id                 UUID (PK)
  email              CITEXT, nullable, partial UNIQUE (active users only)
  password_hash      TEXT, nullable (OAuth-only accounts have none)
  display_name       VARCHAR(50)
  handle             VARCHAR(30), partial UNIQUE (active users only)
  avatar_asset_id    FK -> media_assets.id, nullable
  preferred_locale   VARCHAR(10) (ja / en / fr)
  level, experience_points, rating   INTEGER (Elo, default 1200)
  status             CHECK(active / suspended / deleted)
  anonymized_at      TIMESTAMPTZ, nullable

oauth_accounts
  id                   UUID (PK)
  user_id              FK -> users.id
  provider             CHECK(google / github / 42)
  provider_account_id  VARCHAR(255)
  UNIQUE(provider, provider_account_id)

sessions
  id           UUID (PK)
  user_id      FK -> users.id
  token_hash   TEXT, UNIQUE  (SHA-256 only; raw token never stored)
  expires_at   TIMESTAMPTZ (7 days)
  revoked_at   TIMESTAMPTZ, nullable

matches
  id           UUID (PK)
  mode         CHECK(ranked / casual / ai / friend)
  status       CHECK(in_progress / finished / aborted)
  result_type  CHECK(goal / resign / timeout / draw / abort), nullable
  total_moves  INTEGER
  CHECK(finished implies result_type and finished_at are set)

match_participants
  match_id      FK -> matches.id (composite PK with user_id)
  user_id       FK -> users.id
  seat          CHECK(0 / 1)
  outcome       CHECK(win / loss / draw), nullable while in progress
  rating_before, rating_after, xp_gained   INTEGER

match_actions                     -- append-only move log
  match_id     FK -> matches.id (composite PK with action_seq)
  action_seq   INTEGER  (doubles as the optimistic game version)
  action_id    UUID, UNIQUE(match_id, action_id)  (idempotency key)
  actor_seat   CHECK(0 / 1)
  action_type  CHECK(move / wall / resign / timeout / abort)
  payload      JSONB

friendships
  user_low_id, user_high_id   FKs -> users.id (composite PK, CHECK low < high)
  requested_by_user_id        FK -> users.id
  status                      CHECK(pending / accepted / rejected)

media_assets
  id                UUID (PK)
  owner_user_id     FK -> users.id
  purpose           CHECK(avatar)
  storage_key       TEXT, UNIQUE (random; never exposed via API)
  mime_type, size_bytes, width, height, checksum_sha256
  status            CHECK(active / deleted)

user_achievements
  user_id       FK -> users.id (composite PK with code)
  code          VARCHAR(50)   (definitions live in domain code)
  unlocked_at   TIMESTAMPTZ

api_keys
  id            UUID (PK)
  user_id       FK -> users.id
  key_prefix    VARCHAR(16)  (display only)
  key_hash      CHAR(64), UNIQUE  (SHA-256; raw key shown once at creation)
  scopes        VARCHAR(100)  (read / write)
  expires_at, revoked_at, last_used_at   TIMESTAMPTZ, nullable

blocks
  id                                UUID (PK)
  blocker_user_id, blocked_user_id  FKs -> users.id, UNIQUE pair
```

### Entity Relationships

```
users 1──N oauth_accounts
users 1──N sessions
users 1──N match_participants ──N──1 matches
matches 1──N match_actions
users 1──N friendships        (as low / high side)
users 1──N media_assets       (avatar_asset_id points back to one)
users 1──N user_achievements
users 1──N api_keys
users 1──N blocks             (as blocker / blocked)
```

Generated per-table documentation with ER diagrams lives in
[docs/schema/](docs/schema/README.md) (tbls output, regenerated on schema changes).

## Features List

| Feature                     | Description                                                                      | Member(s)         |
| --------------------------- | -------------------------------------------------------------------------------- | ----------------- |
| Signup & Login              | Email/password (bcrypt) and Google OAuth; separate signup and login screens      | hirwatan          |
| Quoridor Rules Engine       | 9×9 board, 10 walls each, jumps/side-steps, BFS path validation                  | ttanaka           |
| Real-time Match Sync        | Full-state broadcasts, `gameVersion` staleness checks, `actionId` idempotency    | ttanaka           |
| Matchmaking & Timeouts      | Quick-match queue; 60s per move, 45s reconnection grace                          | ttanaka           |
| Rating / XP / Achievements  | Elo (K=32), XP levels, 12 achievements applied transactionally on match end      | sguruge           |
| Statistics Dashboard        | Interactive charts with date/mode/result/opponent filters, SSE live refresh      | sguruge, atashiro |
| Match History & Leaderboard | Dated history with opponents and result types; rating-ordered leaderboard        | sguruge           |
| CSV / PDF Export            | Exports always match the currently applied filters                               | atashiro          |
| Spectator Mode              | Live match list, join mid-game, read-only board, spectator counts                | sguruge           |
| Profiles & Avatars          | Profile pages; content-sniffed png/jpeg/webp uploads ≤ 5MB with default fallback | sguruge           |
| Friend System               | Requests with reciprocal auto-accept, online status (2-minute presence window)   | sguruge           |
| Public API                  | API keys (hashed, scoped, expirable), 60 req/min rate limit, OpenAPI docs        | atashiro          |
| Internationalization        | Complete ja/en/fr translations incl. API errors; cookie-persisted switcher       | kanahash          |
| Shared Modal System         | One modal shell with swappable content (results, confirmations)                  | kanahash          |
| HTTPS / WSS                 | Caddy reverse proxy with local CA certificates                                   | hirwatan          |
| Privacy Policy & ToS        | Substantive legal pages in three languages, linked from every screen             | hirwatan          |

## Modules

<!-- Major = 2pts, Minor = 1pt. Minimum 14 points required. -->

| #   | Category  | Module                                    | Type  | Pts    | Implementation                                                                                    | Member(s) |
| --- | --------- | ----------------------------------------- | ----- | ------ | ------------------------------------------------------------------------------------------------- | --------- |
| 1   | Web       | Use a Framework (FE: Next.js + BE: Gin)   | Major | 2      | Next.js App Router (14 routes, RSC/client split); Gin route tree + middleware chain               | hirwatan  |
| 2   | Web       | Real-time features (WebSocket)            | Major | 2      | Server-authoritative sessions broadcast full snapshots; graceful disconnect/reconnect             | ttanaka   |
| 3   | Web       | Public API (key, rate limit, docs, 5+ EP) | Major | 2      | 8 data endpoints (GET/POST/PUT/DELETE) with scoped hashed keys, 429 + `X-RateLimit-*`, Swagger UI | atashiro  |
| 4   | Gaming    | Complete web-based game (Quoridor)        | Major | 2      | Full rules incl. BFS wall validation; goal/resign/timeout endings; server-side state              | ttanaka   |
| 5   | Gaming    | Remote players                            | Major | 2      | Separate browsers/PCs; latency-safe sync; reconnection grace then forfeit                         | ttanaka   |
| 6   | User Mgmt | Standard user management                  | Major | 2      | Profile pages, validated avatars with default, friends with online status                         | sguruge   |
| 7   | Data      | Advanced analytics dashboard              | Major | 2      | Interactive charts, 4 filters, SSE real-time refresh, CSV/PDF export matching filters             | atashiro  |
| 8   | Web       | Use an ORM                                | Minor | 1      | GORM models drive the Atlas schema; transactional repositories                                    | kanahash  |
| 9   | User Mgmt | Game statistics & match history           | Minor | 1      | Wins/losses/rating/ranking/level/XP, achievements, leaderboard                                    | sguruge   |
| 10  | User Mgmt | OAuth 2.0                                 | Minor | 1      | Google OIDC, constant-time state check, first-login user creation, duplicate-identity prevention  | hirwatan  |
| 11  | Gaming    | Spectator mode                            | Minor | 1      | Live list, mid-game join with latest state, real-time updates, spectator counts                   | sguruge   |
| 12  | A11y/i18n | Multiple languages (3)                    | Minor | 1      | ja/en/fr complete translations, switcher, cookie persistence, ICU plurals, localized dates        | kanahash  |
|     |           | **Total**                                 |       | **19** |                                                                                                   |           |

## Individual Contributions

### hirwatan

- **Role**: Product Owner / Developer
- **Contributions**:
    - Email/password registration and login (bcrypt, timing-equalized failures)
    - Google OAuth 2.0 with CSRF-safe state/nonce handling
    - Migrated routing and middleware from `net/http` to Gin without touching handler internals
    - Caddy HTTPS/WSS setup, legal pages, repository administration and this README
- **Challenges**: Credentialed CORS silently rejects wildcard headers — browsers
  treat `*` literally when cookies are involved; fixed by echoing the preflight's
  requested headers

### ttanaka

- **Role**: Project Manager / Developer
- **Contributions**:
    - Pure-domain Quoridor engine with exhaustive table-driven tests
    - Server-authoritative session layer: optimistic versioning, idempotency keys, turn/grace timers
    - WebSocket transport and the game/board UI with per-player board orientation
- **Challenges**: Making reconnection safe — solved by replaying the append-only
  action log and always broadcasting full snapshots instead of diffs

### sguruge

- **Role**: Tech Lead / Developer
- **Contributions**:
    - Profiles and the validated avatar pipeline (sniffed MIME, decode check, logical delete)
    - Friend system with reciprocal auto-accept and in-memory presence tracking
    - Statistics aggregation, achievements, leaderboard, spectator attachment
- **Challenges**: Keeping stats queries correct once in-progress matches began
  sharing tables with finished ones (NULL outcomes, status filters)

### atashiro

- **Role**: Developer
- **Contributions**:
    - API-key lifecycle: hash-only storage, scopes, expiry and revocation
    - Fixed-window rate limiter with inspectable `X-RateLimit-*` headers
    - OpenAPI coverage for the public surface; analytics opponent filter wired
      through every aggregate and the CSV export
- **Challenges**: Proving rate limits and every key-rejection path with
  reproducible curl demos rather than claims

### kanahash

- **Role**: Developer
- **Contributions**:
    - next-intl integration: cookie persistence, per-locale date/number formatting,
      translation of every screen including error codes
    - Shared modal system (result modal, logout/resign confirmations)
    - ORM/migration hygiene and developer documentation
- **Challenges**: React Server Component boundaries — functions can't cross into
  client components, which shaped several small client wrappers

## Project Management

- **Task Distribution**: GitHub Issues plus a shared module checklist; work split
  into small feature branches with one PR each
- **Code Review**: every change lands through a PR into `main`; an automated
  review workflow plus teammate review with `[must]/[imo]/[nits]/[ask]/[fyi]` prefixes
- **Commit Discipline**: small layered commits (schema → domain → usecase →
  infrastructure → handler → UI), each building on its own
- **Meetings**: weekly sync to re-prioritize modules and unblock work
- **Communication**: Discord server for daily coordination

## Resources

### Documentation & References

- [Quoridor rules (Gigamic)](https://en.gigamic.com/game/quoridor)
- [Next.js Documentation](https://nextjs.org/docs)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/) / [Atlas](https://atlasgo.io/docs)
- [coder/websocket](https://pkg.go.dev/github.com/coder/websocket)
- [next-intl](https://next-intl.dev/docs) / [ICU MessageFormat](https://unicode-org.github.io/icu/userguide/format_parse/messages/)
- [Mantine](https://mantine.dev/) / [Tailwind CSS](https://tailwindcss.com/docs)
- [Google OpenID Connect](https://developers.google.com/identity/openid-connect/openid-connect)
- [bcrypt (Go)](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [Caddy Documentation](https://caddyserver.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

### AI Usage

AI tools (Claude via Claude Code) were used as assistants in the following areas:

- **Code assistance**: scaffolding features (game session layer, API-key
  middleware, i18n sweep) that the responsible member then reviewed and adjusted
- **Testing**: generating table-driven unit tests and reproducible curl/WebSocket
  end-to-end scripts; every claimed behavior was exercised against a running stack
- **Debugging**: framework-specific pitfalls (credentialed CORS wildcards, React
  Server Component prop boundaries, Gin/WebSocket integration)
- **Translations**: English and French message drafts, reviewed for consistency
- **Documentation**: PR descriptions and drafts of this README

All AI-generated code and content was reviewed, understood, tested and validated
by team members before integration; unit tests and end-to-end checks gate every
merge to `main`.

## File Structure

```
42_Transcendence/
├── docker-compose.yml        # Orchestrates all services
├── Makefile                  # docker compose wrapper (make / down / logs / re)
├── .env.example              # Environment variable template
├── README.md                 # This file
├── caddy/Caddyfile           # HTTPS/WSS reverse proxy config
├── docs/schema/              # Generated DB docs (tbls: per-table md + ER diagrams)
│
├── frontend/                 # Next.js web client (see frontend/README.md)
│   ├── app/                  # App Router pages
│   │   ├── login/ signup/    # Auth screens (email+password / Google)
│   │   ├── game/             # Live match screen
│   │   ├── watch/            # Spectator list + live view
│   │   ├── friends/ settings/ users/[userId]/
│   │   ├── stats/ matches/ achievements/ leaderboard/
│   │   └── privacy/ terms/   # Legal pages
│   ├── components/           # Shared UI (board, modals, forms, charts)
│   ├── lib/                  # API helpers, theme, types
│   ├── messages/             # ja / en / fr catalogs (next-intl)
│   └── i18n/                 # Locale resolution
│
└── backend/                  # Go API + WebSocket server (see backend/README.md)
    ├── cmd/serv/             # Entrypoint + Gin middleware (composition root)
    ├── cmd/migrate/          # GORM structs -> DDL (for Atlas diffing)
    ├── domain/               # Entities & rules (Quoridor engine, rating, keys)
    ├── usecase/              # Orchestration + repository interfaces (+ tests)
    ├── infrastructure/       # GORM repositories, presence, rate limiter, files
    ├── handler/              # HTTP/WS handlers (Gin routes via adapter)
    ├── apispec/              # OpenAPI annotation sources
    ├── docs/swagger/         # Generated OpenAPI spec
    └── migrations/           # Versioned SQL migrations (Atlas)
```
