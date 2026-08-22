*This project has been created as part of the 42 curriculum by hirwatan, sguruge, atashiro, ttanaka, kanahash.*

# Transcendence — Online Quoridor Platform

## Description

**Transcendence** is a real-time online platform for playing **Quoridor** — the classic
board game where two players race their pawns across a 9×9 board while placing walls
to slow each other down.

Key features:

- **Real-time matches** between remote players over WebSocket, with a
  server-authoritative rules engine (moves, jumps, wall placement, BFS path
  validation, timeouts, reconnection with a grace period)
- **Quick matchmaking**, Elo rating, XP/levels, achievements, match history,
  statistics dashboard and leaderboard
- **Spectator mode**: watch any live match with live updates and spectator counts
- **Accounts**: email + password (bcrypt) and Google OAuth 2.0 sign-in, profile
  pages, avatars with content validation, friends with online status
- **Public REST API** authenticated by hashed API keys with scopes, rate limiting
  and OpenAPI documentation
- **Internationalization**: full Japanese / English / French translations with a
  language switcher, localized dates, numbers and plurals
- **HTTPS/WSS** everywhere between the browser and the backend (Caddy reverse proxy)

Developer documentation with deeper architectural notes lives in
[`backend/README.md`](backend/README.md) and [`frontend/README.md`](frontend/README.md).

## Instructions

### Prerequisites

- **Docker** (Docker Desktop or an equivalent with Compose v2) and **make**
- A **Google OAuth client** (only needed for Google sign-in; email/password works without it)
- Optional, for running checks outside containers: Go ≥ 1.26, Node ≥ 22 with pnpm, Atlas CLI

### Setup

1. Clone the repository and create your environment file:

   ```bash
   git clone <repository-url> && cd 42_Transcendence
   cp .env.example .env
   ```

2. Fill in the Google OAuth values in `.env` (`GOOGLE_CLIENT_ID`,
   `GOOGLE_CLIENT_SECRET`). In the
   [Google Cloud Console](https://console.cloud.google.com/apis/credentials),
   register the authorized redirect URI (exact match required):

   ```
   https://localhost:8443/auth/google/callback
   ```

3. Build and start everything with a single command:

   ```bash
   make
   ```

4. Open **https://localhost** (frontend) and **https://localhost:8443** (API) once
   each and accept the locally-generated certificate (Caddy issues a self-signed
   local CA for development). Then sign up and play.

Useful commands:

| Command | What it does |
| --- | --- |
| `make` | Build images and start db / backend / frontend / caddy |
| `make down` / `make logs` / `make re` | Stop, tail logs, rebuild from scratch |
| `cd backend && go test ./...` | Backend unit tests (domain rules, sessions, keys) |
| `cd frontend && pnpm lint && pnpm format:check` | Frontend lint / format checks |
| `cd backend && make swagger-serve` | OpenAPI UI at http://localhost:4000/swagger/index.html |

Database migrations (Atlas) are applied automatically when the backend container
starts. Legal pages are served at `/privacy` and `/terms`.

## Team Information

| Login | Role(s) | Responsibilities |
| --- | --- | --- |
| **hirwatan** | Product Owner / Developer | Product direction and backlog, validation of completed work; authentication (email+password, OAuth 2.0), framework migration (Gin), HTTPS/Caddy, repository administration |
| **ttanaka** | Project Manager / Developer | Planning, task tracking and coordination; Quoridor rules engine, real-time synchronization, remote play and reconnection logic |
| **sguruge** | Technical Lead / Developer | Architecture (Clean Architecture layering) and code review; user management, friends & presence, statistics, achievements, spectator mode |
| **atashiro** | Developer | Public API with API keys and rate limiting, analytics dashboard filters and exports |
| **kanahash** | Developer | ORM data layer, internationalization (ja/en/fr), shared UI (modals, auth screens) |

## Project Management

- **Task organization**: GitHub Issues for planned work and a shared task sheet for
  module checklists; work was split into small feature branches with one PR each.
- **Code review**: every change went through a pull request into `main`; an
  automated review workflow plus at least one teammate reviewed significant PRs.
  Commits are small and layered (schema → domain → usecase → infrastructure →
  handler → UI) so each commit builds on its own.
- **Communication**: a team Discord server for daily coordination, plus PR
  comments (with `[must]/[imo]/[nits]/[ask]/[fyi]` prefixes) for review discussion.
- **Cadence**: weekly sync to re-prioritize modules and unblock work.

## Technical Stack

| Area | Choice | Why |
| --- | --- | --- |
| Frontend | **Next.js 16** (App Router) + TypeScript | File-based routing and the Server/Client component split fit our screens; SSR keeps first paint fast |
| UI | **Mantine** + **Tailwind CSS v4** | Ready-made accessible components combined with utility styling; they coexist through CSS cascade layers |
| i18n | **next-intl** | Cookie-based locale without URL changes; ICU messages give plurals and localized dates/numbers |
| Backend | **Go 1.26** + **Gin** | Goroutines handle many concurrent WebSocket connections; Gin provides the route tree (path params, groups) and the middleware chain (recovery → access log → CORS) |
| Architecture | **Clean Architecture** | Game rules live in `domain` with zero HTTP/DB imports, so the whole rules engine is unit-testable |
| Real-time | **coder/websocket** (WS) + SSE | Server-authoritative game state pushed as full snapshots; SSE for statistics refresh |
| Database | **PostgreSQL 16** | Relational integrity for match/participant data; CHECK constraints and partial unique indexes encode invariants |
| ORM / migrations | **GORM** + **Atlas** | Struct-driven schema with versioned SQL migrations; transactions guard multi-row updates |
| Auth | Cookie sessions + **bcrypt** + Google **OAuth 2.0 / OIDC** | Hashed+salted passwords; opaque session tokens stored only as SHA-256 hashes |
| HTTPS | **Caddy** | One small container terminates TLS for the frontend and the API (including WSS) with a local CA |

Framework usage in detail (where, not just installed):

- **Next.js**: 14 routes under `frontend/app/` (game, watch, friends, settings,
  stats, matches, achievements, leaderboard, auth, legal…); Server Components
  fetch and gate auth, Client Components own the board, forms and WebSocket hook.
- **Gin**: `backend/handler/handler.go` builds every route on `gin.Engine`
  (route groups, `/users/:userId` beside `/users/me`); `backend/cmd/serv/main.go`
  chains `gin.Recovery()`, an access logger and CORS as middleware.

## Database Schema

PostgreSQL with 10 tables. Generated, always-up-to-date docs (per-table Markdown +
ER diagrams) live in [`docs/schema/`](docs/schema/README.md) (`tbls` output).

Core relations:

- `users` — profile, bcrypt `password_hash` (nullable for OAuth-only), rating, XP;
  partial unique indexes free handles/emails of anonymized accounts
- `oauth_accounts` — `(provider, provider_account_id)` unique: one Google identity
  maps to exactly one user
- `sessions` — SHA-256 token hashes only, 7-day expiry
- `matches` + `match_participants` — one row per match, two participant rows
  (seat 0/1) carrying outcome and rating before/after
- `match_actions` — append-only move log; `action_seq` doubles as the optimistic
  game version, `(match_id, action_id)` unique gives idempotent retries
- `friendships` (normalized `user_low_id < user_high_id`), `media_assets`
  (avatars; hashed storage keys), `user_achievements`, `api_keys` (hashed keys,
  scopes, expiry/revocation)

## Features List

| Feature | Owner(s) | Notes |
| --- | --- | --- |
| Quoridor rules engine (moves, jumps, walls, BFS path check) | ttanaka | Pure `domain` package, table-driven tests |
| Real-time match sync, reconnection, idempotent actions | ttanaka | Full-state broadcasts; `gameVersion` + `actionId` |
| Quick matchmaking, turn/disconnect timeouts | ttanaka | 60s per move, 45s reconnect grace |
| Elo rating, XP/levels, achievements | sguruge | Applied transactionally on match end |
| Statistics dashboard, match history, leaderboard, CSV/PDF export | sguruge / atashiro | Filters: date range, mode, result, opponent |
| Spectator mode with live counts | sguruge | Watch any in-progress match mid-game |
| Profiles, avatars (validated uploads), friends, online status | sguruge | Content-sniffed png/jpeg/webp ≤ 5MB |
| Email+password and Google sign-in, separate signup/login screens | hirwatan | bcrypt; timing-safe login errors |
| Public API with API keys, scopes, rate limiting, OpenAPI | atashiro | 8 data endpoints (GET/POST/PUT/DELETE) |
| Internationalization ja/en/fr, language switcher | kanahash | ICU plurals, localized dates, translated API errors |
| Shared modal system (result, confirmations) | kanahash | One shell, swappable content |
| HTTPS/WSS via Caddy, legal pages | hirwatan | Local CA certificates |

## Modules

Target: 14 points. Claimed: **7 Major (14 pts) + 5 Minor (5 pts) = 19 points.**

| # | Module | Type | Pts | Owner | Implementation summary |
| --- | --- | --- | --- | --- | --- |
| 1 | Framework for frontend and backend | Major | 2 | hirwatan | Next.js App Router on the frontend; Gin routing + middleware on the backend (see Technical Stack) |
| 2 | Real-time features (WebSocket) | Major | 2 | ttanaka | Server-authoritative sessions broadcast full board snapshots; graceful disconnect/reconnect; stale ops rejected via `gameVersion`, retries deduplicated via `actionId` |
| 3 | Public API (API key, rate limit, docs, 5+ endpoints) | Major | 2 | atashiro | `/v1` endpoints (GET/PUT profile, GET matches/stats/leaderboard/friends, POST/DELETE friends) with read/write scopes; raw key shown once, SHA-256 stored; 60 req/min fixed window with `X-RateLimit-*`; Swagger UI |
| 4 | Complete web-based game (Quoridor) | Major | 2 | ttanaka | 9×9 board, 10 walls each, jump/side-step rules, BFS validates every wall keeps a path; goal/resign/timeout endings |
| 5 | Remote players | Major | 2 | ttanaka | Separate browsers/PCs play live; latency-safe full-state sync; reconnection grace, then forfeit |
| 6 | Standard user management | Major | 2 | sguruge | Profile view/edit pages, validated avatar upload with default fallback, friends with online status |
| 7 | Advanced analytics dashboard | Major | 2 | atashiro | Interactive charts (line/bar/donut), filters (dates, mode, result, opponent), SSE refresh on match end, CSV/PDF export matching current filters |
| 8 | ORM | Minor | 1 | kanahash | GORM models drive the Atlas schema; repositories use transactions for multi-row updates |
| 9 | Game statistics / match history | Minor | 1 | sguruge | Wins/losses/rating/ranking/level/XP, dated history with opponents and result types, achievements, leaderboard |
| 10 | OAuth 2.0 | Minor | 1 | hirwatan | Google OIDC with state/nonce verification (constant-time), first-login user creation, duplicate-identity prevention |
| 11 | Spectator mode | Minor | 1 | sguruge | Live match list, join mid-game with latest state, real-time updates, read-only, spectator counts |
| 12 | Multiple languages | Minor | 1 | kanahash | ja/en/fr complete translations incl. API errors; switcher persists via cookie; ICU plurals and localized dates |

## Individual Contributions

- **hirwatan** — Authentication end to end: email+password registration/login
  (bcrypt, timing-equalized failures), Google OAuth with CSRF-safe state
  handling; migrated routing/middleware from `net/http` to Gin without touching
  handler internals; Caddy HTTPS/WSS; legal pages; this README. Challenge:
  credentialed CORS silently rejects wildcard headers — fixed by echoing the
  preflight's requested headers.
- **ttanaka** — The game itself: pure-domain Quoridor engine with exhaustive
  table-driven tests; server-authoritative session layer (optimistic versioning,
  idempotency keys, timers); WebSocket transport; game/board UI with per-player
  board orientation. Challenge: making reconnection safe — solved by replaying
  the append-only action log and broadcasting full snapshots.
- **sguruge** — User-facing platform: profiles and validated avatar pipeline,
  friends with reciprocal auto-accept and in-memory presence, statistics
  aggregation, achievements, spectator attachment that never touches player
  timers. Challenge: keeping stats queries correct once in-progress matches
  shared tables with finished ones.
- **atashiro** — Public surface: API-key lifecycle (hash-only storage, scopes,
  expiry/revocation), fixed-window rate limiter with inspectable headers,
  OpenAPI coverage; analytics opponent filter wired through every aggregate and
  the CSV export. Challenge: proving rate limits and key rejection paths with
  reproducible curl demos.
- **kanahash** — Cross-cutting UX: next-intl integration with cookie persistence
  and per-locale formatting, translation of every screen including error codes;
  the shared modal system; ORM/migration hygiene. Challenge: React Server
  Component boundaries — functions can't cross into client components, which
  shaped several small client wrappers.

## Resources

- Quoridor rules — Gigamic: https://en.gigamic.com/game/quoridor
- Next.js App Router: https://nextjs.org/docs
- Gin: https://gin-gonic.com/docs/ / Go net/http: https://pkg.go.dev/net/http
- GORM: https://gorm.io/docs/ / Atlas migrations: https://atlasgo.io/docs
- coder/websocket: https://pkg.go.dev/github.com/coder/websocket
- next-intl: https://next-intl.dev/docs / ICU MessageFormat: https://unicode-org.github.io/icu/userguide/format_parse/messages/
- Mantine: https://mantine.dev/ / Tailwind CSS: https://tailwindcss.com/docs
- OAuth 2.0 / OIDC: https://developers.google.com/identity/openid-connect/openid-connect
- bcrypt: https://pkg.go.dev/golang.org/x/crypto/bcrypt
- Caddy: https://caddyserver.com/docs/
- tbls (schema docs): https://github.com/k1LoW/tbls

### How AI was used

AI assistance (Claude, via Claude Code) was used throughout the project as a
productivity tool, always followed by human review:

- **Implementation scaffolding**: first drafts of features (game session layer,
  API-key middleware, i18n sweep) which the responsible member then reviewed,
  adjusted and tested.
- **Tests and verification**: generating table-driven unit tests and reproducible
  curl/WebSocket E2E scripts; every claimed behavior in our PRs was exercised
  against a running stack.
- **Translations**: English and French message drafts, reviewed for tone and
  consistency.
- **Debugging**: diagnosing framework-specific pitfalls (credentialed CORS
  wildcards, React Server Component prop boundaries, Gin/WebSocket integration).
- **Documentation**: PR descriptions and drafts of this README.

All AI-generated code was read, understood and tested by the team member who
shipped it; unit tests and end-to-end checks gate every merge to `main`.
