# Frontend — TypeScript

Next.js 16 (App Router) / React 19 / TailwindCSS v4 + Mantine v9. Package manager: pnpm.

## Running

```bash
# Docker (from the repository root)
make up                 # https://localhost (via Caddy) / http://localhost:3000 (direct)
make exec-frontend

# Directly on the host
corepack enable pnpm            # enables the pnpm version pinned in package.json
pnpm install --frozen-lockfile
pnpm dev                        # :3000
```

Prerequisites: Node.js ≥ 22.13 / pnpm 11.17.0 (see `engines` and `packageManager` in `package.json`).

| Command             | What it does                          |
| ------------------- | ------------------------------------- |
| `pnpm dev`          | Dev server (hot reload)               |
| `pnpm build`        | Production build                      |
| `pnpm start`        | Production server (requires `build`)  |
| `pnpm lint`         | Static analysis with ESLint           |
| `pnpm lint:fix`     | ESLint auto-fix                       |
| `pnpm format`       | Format with Prettier                  |
| `pnpm format:check` | Detect unformatted files (same as CI) |

> In Docker, `./frontend` is bind-mounted so hot reload works, but `node_modules`
> is masked by an anonymous volume. **After adding a dependency, rebuild the image
> and discard the anonymous volume:**
>
> ```bash
> docker compose build frontend
> docker compose up -d -V frontend   # without -V the stale node_modules survives
> ```
>
> If you forget `-V`, pnpm inside the container notices the mismatch, tries to
> reinstall, and fails to start with `ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`.

## How the App Router works

**The directory structure under `app/` is the URL.** There is no routing config file.

```
app/
├── layout.tsx          →  shared shell for every page (required)
├── page.tsx            →  /
├── game/
│   └── page.tsx        →  /game
└── users/
    └── [userId]/
        └── page.tsx    →  /users/123  ([userId] is a dynamic segment)
```

File names have fixed roles:

| File            | Role                                                              |
| --------------- | ----------------------------------------------------------------- |
| `page.tsx`      | The screen for that path; the URL only exists once this file does |
| `layout.tsx`    | Wrapper around descendant pages; not remounted on navigation      |
| `loading.tsx`   | Shown automatically while data loads                              |
| `error.tsx`     | Shown automatically on errors (must be a Client Component)        |
| `not-found.tsx` | Shown for 404s                                                    |
| `route.ts`      | An API endpoint instead of a screen                               |

Only the root layout ([app/layout.tsx](app/layout.tsx)) writes `<html>` and `<body>`.
Exporting `metadata` (or `generateMetadata`) produces `<head>`; a page can export its
own to override it. Fonts are bundled at build time via `next/font`, so no request
goes to Google at runtime.

## Server Components and Client Components

In the App Router **every component is a Server Component by default** — the
biggest difference from the Pages Router.

|                            | Server Component (default) | Client Component (`"use client"`) |
| -------------------------- | -------------------------- | --------------------------------- |
| Where it runs              | Server only                | Server (first HTML) + browser     |
| `useState` / `useEffect`   | Not available              | Available                         |
| Event handlers (`onClick`) | Not available              | Available                         |
| DB / secret access         | Allowed (never leaks)      | Not allowed                       |
| JS bundle size             | Free                       | Adds to the bundle                |

```tsx
// Server Component (default): can be async and await directly
export default async function Page() {
  const res = await fetch("http://backend:4000");
  const data = await res.json();
  return <p>{data.message}</p>;
}
```

```tsx
"use client"; // ← first line of the file makes it a Client Component

import { useState } from "react";

export default function Counter() {
  const [n, setN] = useState(0);
  return <button onClick={() => setN(n + 1)}>{n}</button>;
}
```

Guidelines:

1. **Write Server Components first.**
2. When you need state, event handlers or browser APIs (`window`, `localStorage`),
   extract **just that part** into a small Client Component.
3. Avoid putting `"use client"` on a whole page — everything below it ships to the client.
4. Functions cannot cross the RSC boundary: passing `component={Link}` from a
   Server Component into a Mantine component throws at render time. Wrap it in a
   tiny Client Component instead (see `components/auth-link.tsx`).

## Directory layout

**Only `app/`, `components/` and `lib/`.** When unsure, ask: “does it become a URL /
is it used by more than one screen?”

```
frontend/
├── app/                  routing and screens (App Router)
├── components/           UI parts shared by two or more screens
├── lib/                  shared non-UI logic (theme, API URLs, request helper)
├── messages/             ja / en / fr translation catalogs (next-intl)
├── i18n/                 locale config and request resolver
└── public/               static files
```

Placement rules:

| What you are writing            | Where it goes                         |
| ------------------------------- | ------------------------------------- |
| A screen (has a URL)            | **Directly in** `app/<path>/page.tsx` |
| A part used only by that screen | Inside/next to that `page.tsx`        |
| A part used by 2+ screens       | `components/`                         |
| Non-UI logic, constants, types  | `lib/`                                |

- **Don't move screens out of `app/`.** Keep the screen body in `page.tsx` rather
  than making it a thin wrapper, so anyone can go from URL to code in one hop.
- **Don't start in `components/`.** Move a part there when a second screen needs
  it; premature extraction accumulates parts that were never actually shared.
- File names are kebab-case; exports are named (only `page.tsx` default-exports).
- When flat `components/` starts to hurt, cut subdirectories like
  `components/game/` at that moment — never create empty directories in advance.

`@/*` maps to the project root (`tsconfig.json` paths), so no `../../` imports:

```tsx
import { apiUrl } from "@/lib/api";
import { QuoridorMark } from "@/components/quoridor-mark";
```

## Types and schema

API response types are declared next to their consumers (e.g. `lib/stats.ts`,
`lib/auth.ts`, socket types in `components/use-game-socket.ts`), each mirroring the
backend handler DTO it corresponds to — the comment on each type names its Go
counterpart. Keep the two in sync by hand when a handler changes.

- These are **API response shapes, not table definitions**. Server-only fields
  (`password_hash`, `token_hash`, `storage_key`) never appear in them.
- Timestamps arrive as `string` over JSON; convert to `Date` at the consumer if needed.
- The authoritative database schema is the generated documentation in
  [docs/schema/](../docs/schema/README.md) (tbls output per table with ER diagrams).

## TailwindCSS v4 + Mantine v9

**Tailwind is the styling substrate; Mantine supplies the components.**
Don't hand-roll primitives (buttons, cards, alerts, modals) — import them from
[@mantine/core](https://mantine.dev/), copy ready-made blocks from
[ui.mantine.dev](https://ui.mantine.dev/) into `app/`/`components/`, then adjust
spacing/layout with Tailwind utilities.

### Layer order is everything

The whole coexistence hangs on the first line of [app/globals.css](app/globals.css).
**Later layers win**, so this order is the authority on precedence:

```css
@layer theme, base, mantine, components, utilities;

@import "tailwindcss/theme.css" layer(theme);
@import "tailwindcss/preflight.css" layer(base);
@import "@mantine/core/styles.layer.css"; /* internally wrapped in @layer mantine */
@import "tailwindcss/utilities.css" layer(utilities);
```

- `mantine` after `base` (Tailwind's reset) → preflight can't break Mantine's look
- `utilities` after `mantine` → **`<Paper className="p-0">` works without `!important`**
- Import **`styles.layer.css`**, not `@mantine/core/styles.css` — the latter is
  unlayered and escapes ordering control
- Don't use the all-in-one `@import "tailwindcss"` either; it prevents inserting
  Mantine **between** base and utilities

### Colors and tokens

Colors exist in exactly one place: [lib/theme.ts](lib/theme.ts) (`createTheme`).
`@theme inline` in `globals.css` bridges them to Tailwind class names, so both
worlds resolve to the same values:

```tsx
<main className="bg-body">        {/* → var(--mantine-color-body) */}
<p className="text-dimmed">       {/* → var(--mantine-color-dimmed) */}
<span className="bg-emerald-500"> {/* → var(--mantine-color-emerald-5) */}
```

- **Never write raw colors.** `bg-[#0f172a]` and `text-slate-400` are banned, so a
  palette change stays a one-file edit in `theme.ts`.
- To add a color: add a 10-step tuple in `theme.ts` and matching lines in
  `@theme inline`.
- The scheme is fixed to dark (`forceColorScheme="dark"` in `layout.tsx`).

### Which tool for what

| Goal                                          | Use                                      |
| --------------------------------------------- | ---------------------------------------- |
| Buttons, inputs, cards, modals, notifications | Mantine components                       |
| Spacing, alignment, grids, responsiveness     | Tailwind utilities                       |
| Anything a Mantine prop covers                | Mantine props (`gap` `c` `maw` `size`)   |
| Complex styles beyond props and Tailwind      | A colocated `*.module.css` (last resort) |

`*.module.css` files can use `rem()` and `$mantine-breakpoint-*`
([postcss.config.mjs](postcss.config.mjs)). Never write bare `px`; go through `rem()`.

### Icons

Use [@tabler/icons-react](https://tabler.io/icons), Mantine's assumed icon set.
Only product-specific marks (logo, Google G) live in `components/*-mark.tsx`.

## Internationalization

next-intl with cookie-based locale (no URL prefix). `i18n/request.ts` resolves the
locale from the `NEXT_LOCALE` cookie (default `ja`); catalogs live in
`messages/{ja,en,fr}.json`. Rules:

- Every user-facing string goes through `useTranslations` (client) or
  `getTranslations` (server). No hardcoded copy.
- Counts use ICU plurals; dates and numbers go through `useFormatter`.
- API/WebSocket errors are translated on the client keyed by the server's
  machine-readable `code`, falling back to the server message.

## Calling the backend

The URL differs by caller:

| Caller                           | URL                                                                             |
| -------------------------------- | ------------------------------------------------------------------------------- |
| Client Component / browser       | `process.env.NEXT_PUBLIC_API_URL` (default `https://localhost:8443`, via Caddy) |
| Server Component / Route Handler | `API_URL` (`http://backend:4000`, container-to-container)                       |

Only variables prefixed `NEXT_PUBLIC_` are embedded in the browser bundle —
so **never put secrets behind a `NEXT_PUBLIC_` name**.

Allowed origins are listed in `backend/cmd/serv/main.go` (`allowedOrigins` for CORS,
`allowedWSOrigins` for WebSocket handshakes). See the root
[README](../README.md) for the service topology and HTTPS setup.

## Lint / Format

Responsibilities are split:

- **Prettier** ([.prettierrc.json](.prettierrc.json)) — formatting only
- **ESLint** ([eslint.config.mjs](eslint.config.mjs)) — bug/convention detection only

`eslint-config-prettier` sits **last** in the ESLint config to disable stylistic
rules, so the two never fight.

```bash
pnpm format      # format
pnpm lint        # static analysis
pnpm lint:fix    # apply auto-fixes
```

CI: [frontend-lint-format.yml](../.github/workflows/frontend-lint-format.yml) runs on
PRs/pushes touching `frontend/**`. Run `pnpm format && pnpm lint` before pushing.

Review checklist: [pull_request_template.md](../.github/pull_request_template.md).
