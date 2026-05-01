# jaycestuff

Jayce Bordelon's production monorepo. All services are deployed to a single Digital Ocean droplet running Docker Compose behind Traefik as a reverse proxy with automatic Let's Encrypt TLS.

## Architecture

**Single-server monolithic deployment.** Traefik routes incoming HTTPS requests to the correct container by hostname and path:

- `jaycebordelon.com` / `www.jaycebordelon.com` → Next.js portfolio (port 3000)
- `auth.jaycebordelon.com` → Go centralized OAuth identity provider (port 8081)
- `vibetradez.com` → `/api/*`, `/auth/*`, `/health` → Go API server (port 8080, priority 20)
- `vibetradez.com` → everything else → Next.js trading frontend (port 3001, priority 10)

## Project Structure

```
jaycestuff/
├── jaycebordelon.com/           # Personal portfolio & blog
│   ├── app/                     # Next.js 16 App Router pages
│   ├── components/              # React components + shadcn/ui
│   ├── content/                 # MDX blog posts
│   ├── lib/                     # Utilities
│   ├── Dockerfile               # Multi-stage Node.js build
│   └── package.json             # Next.js 16, React 19, Tailwind v4
│
├── auth.jaycebordelon.com/      # Centralized OAuth identity provider
│   ├── cmd/server/              # Main entry point
│   ├── internal/
│   │   ├── config/              # Environment variable loading (fail-fast)
│   │   ├── google/              # Google OAuth client (upstream IdP)
│   │   ├── server/              # /oauth/{authorize,token,verify,revoke} + /auth/google/*
│   │   └── store/               # PostgreSQL: users, sessions, oauth_clients, oauth_codes, access_tokens
│   ├── Dockerfile               # Multi-stage Go build
│   └── go.mod
│
├── vibetradez.com/
│   ├── server/                  # Go API server (trading backend)
│   │   ├── cmd/scanner/         # Main entry point, cron jobs, workflows
│   │   ├── internal/
│   │   │   ├── config/          # Environment variable loading
│   │   │   ├── email/           # Resend email client
│   │   │   ├── schwab/          # Schwab OAuth + Market Data API
│   │   │   ├── sentiment/       # Market signal aggregator (StockTwits, Yahoo, Finviz, EDGAR)
│   │   │   ├── server/          # HTTP API handlers
│   │   │   ├── store/           # PostgreSQL database layer
│   │   │   ├── templates/       # HTML email templates
│   │   │   └── trades/          # OpenAI trade analysis + prompts
│   │   ├── Dockerfile           # Multi-stage Go build
│   │   └── go.mod
│   │
│   └── client/                  # Next.js trading frontend
│       ├── app/                 # App Router: /, /history, /terms, /faq
│       ├── components/          # Dashboard, history, layout, subscribe
│       ├── hooks/               # Custom React hooks (live quotes, etc.)
│       ├── lib/                 # API client, formatters, calculations
│       ├── types/               # TypeScript interfaces
│       ├── Dockerfile           # Multi-stage Node.js build
│       └── package.json         # Next.js 16, React 19, shadcn/ui, Recharts
│
├── .github/workflows/          # CI/CD pipeline
│   ├── main-pipeline.yml       # Orchestrator: parallel lint, single build, single deploy, single notify
│   ├── sync.yml                # Git pull on production server
│   ├── lint-portfolio.yml      # Biome lint for jaycebordelon.com
│   ├── lint-trading-frontend.yml # Biome lint for vibetradez.com/client
│   ├── lint-trading-server.yml # golangci-lint for vibetradez.com/server (per-job TMPDIR + GOLANGCI_LINT_CACHE)
│   ├── lint-auth-service.yml   # golangci-lint for auth.jaycebordelon.com (per-job TMPDIR + GOLANGCI_LINT_CACHE)
│   ├── build.yml               # Docker compose build (all three images in one shot)
│   ├── deploy.yml              # Unified rolling deployment: portfolio + trading in one job with continue-on-error per side
│   ├── cleanup.yml             # Post-deploy docker system prune
│   ├── healthcheck.yml         # Endpoint verification + granular /health
│   ├── notify.yml              # Consolidated deploy email (slate theme, per-site status + trading /health)
│   └── cd.yml                  # Standalone manual trigger pipeline
│
├── docker-compose.yml          # All services + Traefik config
└── CLAUDE.md                   # This file
```

## Tech Stack

| Project | Stack |
|---------|-------|
| jaycebordelon.com | Next.js 16, React 19, Tailwind CSS v4, shadcn/ui (new-york), MDX, Framer Motion |
| vibetradez.com/client | Next.js 16, React 19, Tailwind CSS v4, shadcn/ui (new-york), Recharts v3, TradingView Lightweight Charts |
| vibetradez.com/server | Go 1.25, PostgreSQL (Digital Ocean managed), Anthropic Claude Opus 4.7, Schwab Market Data API, Resend email |
| auth.jaycebordelon.com | Go 1.25, PostgreSQL (Digital Ocean managed), golang.org/x/oauth2, bcrypt |
| Infrastructure | Docker Compose, Traefik v2.10, Let's Encrypt, Digital Ocean Droplet |

## Database

PostgreSQL hosted on Digital Ocean Managed Databases. Two DBs: `vibetradez` (trading server) and `auth` (identity provider). Each Go service owns its own DB and auto-migrates schema on startup (CREATE TABLE IF NOT EXISTS).

## Environment Variables

Each service reads **its own per-service `.env` file** (mounted into its container via `env_file:` in `docker-compose.yml`), not a shared root `.env`. Every service also ships a `.env.example` next to its code. Required vars cause the service to `log.Fatal` on boot if missing — a misconfigured container will never serve traffic.

- `auth.jaycebordelon.com/.env` — see `auth.jaycebordelon.com/.env.example`
- `vibetradez.com/server/.env` — see `vibetradez.com/server/.env.example`
- `vibetradez.com/local/.env` — optional overrides for local-dev compose

## Development Rules

### Always lint AND build before pushing

Run these checks before every push. No exceptions. CI will fail if they don't pass, and a failed pipeline blocks deployment for everyone.

```bash
# Lint Go
cd vibetradez.com/server && gofmt -w . && go vet ./...

# Lint + auto-format Next.js (both projects, run from jaycebordelon.com/ where biome is installed).
# CRITICAL: use --write so formatter issues are FIXED, not just reported. CI runs biome
# with format-as-error semantics, so a stray unwrapped string or long line will fail
# Lint + Build • Trading Frontend even though `biome check` (no --write) only "warns".
cd jaycebordelon.com && npx biome check --write .
cd jaycebordelon.com && npx biome check --write ../vibetradez.com/client/

# Build Next.js (both projects)
cd jaycebordelon.com && npx next build
cd vibetradez.com/client && npx next build
```

If any lint or build fails, fix it before pushing. Never push code that hasn't been verified locally. **`biome check` without `--write` is not enough** — it reports format errors but doesn't apply them, so the working tree still has CI-failing files when you commit. Always use `--write`.

### `gh` commands target the wrong repo by default

The personal worktree (`~/Career/jaycestuff/`) shares a machine with work worktrees (`~/Mechanize/Code/eden-*`). When `gh pr checks` / `gh run view` / `gh api` is run from a directory whose git remote points elsewhere — or when the working directory got reset back to a Mechanize worktree — `gh` will silently query the wrong repo and return `HTTP 404` against `mechanize-work/eden`.

**Always pass `--repo JayceBordelon/jaycestuff` explicitly** for any `gh` command relating to this project, or `cd ~/Career/jaycestuff && gh ...`. Don't trust the working directory to be the personal repo just because the last command was.

### Always read the latest documentation

When working with Next.js, shadcn/ui, Tailwind CSS, Recharts, or any external library, **always fetch and read the current documentation** before writing code. Do not rely on recalled syntax or API signatures — they may be outdated. This applies even if it takes extra time. Incorrect assumptions about APIs cause more rework than the time saved by skipping docs.

### Recharts (currently pinned at v3)

`vibetradez.com/client` uses **Recharts ^3.8.0** wrapped by the shadcn `ChartContainer` primitive at `components/ui/chart.tsx`. Recharts 3 was a hard break from 2 — read the migration guide before touching any chart code.

**Reference URLs:**

- v2 → v3 migration guide: <https://github.com/recharts/recharts/wiki/3.0-migration-guide>
- Release notes (changelog after 2.x lives only here): <https://github.com/recharts/recharts/releases>
- npm: <https://www.npmjs.com/package/recharts>

**v3 breaking changes that bite us in this codebase:**

- `CategoricalChartState` is gone. Anything that used to read internal chart state via `Customized` or props now must use hooks (`useActiveTooltipLabel`, etc.).
- Many "internal" cloned props are gone: `Scatter.points`, `Area.points`, `Legend.payload`, `activeIndex`. If you see code reading any of these, it's broken on v3.
- `<Customized />` no longer receives extra props.
- `ref.current.current` on `ResponsiveContainer` is gone.
- `XAxis` / `YAxis` axis lines now render even when there are no ticks.
- Multiple `YAxis` instances render in alphabetical order of `yAxisId`, not render order.
- `CartesianGrid` requires explicit `xAxisId` / `yAxisId` to match the axes it pairs with.
- SVG z-order is the JSX render order — to put a series on top, render it last.
- `Area`'s `connectNulls=true` now treats null datapoints as zero instead of skipping them.
- `Pie.blendStroke` is removed; use `stroke="none"`.
- `<Cell>` is **deprecated** as of v3.7 and will be removed in v4. Migrate per-bar/per-slice colors to the chart element's `shape` prop instead. We still use `Cell` in `daily-pnl-chart.tsx` and `daily-breakdown.tsx` — leave them alone for now but plan a migration before bumping major.
- Tooltip custom-content prop type is now `TooltipContentProps`, not `TooltipProps`.
- Since v3.3, every chart accepts a `responsive` prop directly, so `ResponsiveContainer` wrapping is **optional**. Our shadcn `ChartContainer` still wraps with `ResponsiveContainer` for the inline-style fallback.

**Project-specific rules for chart components:**

- Always render charts through `ChartContainer` from `@/components/ui/chart` — it owns the `ResponsiveContainer`, the `--color-*` CSS variable injection, and the tooltip context.
- Never call `.map()` directly on a `data` prop you receive from a parent without a fallback. The `Cannot read properties of null (reading 'map')` runtime crash on `/history` was caused by the server returning `{"days": null}` for an empty range and `filterByRank` calling `data.days.map(...)` unguarded. The lesson: any boundary that produces JSON arrays must initialize them as empty slices server-side (Go nil slice → JSON `null`), and any client function that consumes them must `?? []` them defensively. Same pattern applies to `cmd/scanner/main.go` and any future endpoint that returns lists.
- When passing data into Recharts components, the data prop must be an array, not null/undefined. A guard like `data && data.length > 0 && <BarChart data={data} ...>` is the safest pattern.

### Always use feature branches

Never push directly to `main`. Create a descriptive branch, push there, and let the user handle PRs and merging.

### Design system consistency

Both Next.js frontends share the same design tokens (CSS variables in `globals.css`), font stack (Plus Jakarta Sans, JetBrains Mono), and shadcn/ui configuration (new-york style, neutral base color, lucide icons). Any UI changes must be consistent across both sites.

## API Protection

All `/api/*` routes on the trading server require the `X-VT-Source` header. Without it, requests return 403. The Next.js frontend includes this header on every fetch call.

## Centralized Auth

`auth.jaycebordelon.com` is a standalone Go OAuth identity provider. It owns the Google OAuth dance (it's the only service Google Cloud knows about as a redirect target) and issues opaque access tokens to registered consumer apps over the authorization-code flow.

- Consumers register via `OAUTH_CLIENTS_JSON` on the auth service and hold a matching client_id + secret.
- Sign-in flow: consumer app redirects browser to `/auth/sso/start` → trading-server generates CSRF state (cookie-scoped to `/auth/sso`) → redirects to `auth.jaycebordelon.com/oauth/authorize` → Google consent → auth-service issues a one-shot code → consumer callback exchanges the code for an access token at `/oauth/token` → consumer sets its own session cookie (`vt_session` for vibetradez) holding the opaque token.
- Per-request verification: consumers call `POST /oauth/verify` on the auth service (cached 60s) to resolve a token into a user. Revocation propagates within the cache TTL.
- Logout revokes via `POST /oauth/revoke`.

Cross-apex cookie sharing (vibetradez.com ↔ jaycebordelon.com) is intentionally not used; each consumer holds its own same-site session cookie and talks to the auth service over HTTP.

## Trading Server Workflows

The Go server runs three cron jobs in Eastern Time:
- **9:25 AM Mon-Fri** — Aggregate market signals (StockTwits, Yahoo Finance, Finviz, SEC EDGAR), call Claude for ranked trade picks, save to DB, email subscribers
- **4:05 PM Mon-Fri** — Fetch closing prices from Schwab, compute EOD P&L, save summaries, email subscribers
- **4:30 PM Fridays** — Aggregate weekly performance, compute stats (win rate, Sharpe, drawdown), email weekly report

Market holidays are hardcoded in `cmd/scanner/main.go`. Jobs skip on holidays and weekends.

## Common Operations

### Re-authorize Schwab OAuth
Visit `https://vibetradez.com/auth/schwab` in a browser. Tokens are stored in the `oauth_tokens` table and auto-refresh.

### Check server health
```bash
curl https://vibetradez.com/health | jq
```
Returns per-service status for database, Anthropic, Schwab, and API with latencies. The Anthropic check goes through the official SDK and warns (instead of fails) when a stub local key is detected.

### Docker commands on production
```bash
ssh jayce@<server>
cd ~/jaycestuff
docker compose logs trading-server --tail 50    # View Go server logs
docker compose logs trading-frontend --tail 50  # View Next.js logs
docker compose restart trading-server           # Restart Go server
docker compose up -d --force-recreate trading-server  # Full recreate
```

## Trade Analysis

The morning trade pipeline uses **a single language model**: Anthropic Claude.

1. **Claude (Opus 4.7 by default)** generates 10 ranked trade ideas via `vibetradez.com/server/internal/trades/picker.go`. The picker uses the official `github.com/anthropics/anthropic-sdk-go` SDK with multi-round Schwab `get_stock_quotes` / `get_option_chain` tools and built-in `web_search`. Each trade comes back with a 1-10 conviction `score` and a free-form `rationale` defending the score.
2. The same picker handles end-of-day analysis via `GetEndOfDayAnalysis`: morning trades are passed back in, Claude fetches closing Schwab marks via the same toolset, and returns realised entry/closing prices plus a brief notes string per pick.
3. `cmd/scanner/main.go` saves the picks to the `trades` table, fires the morning email, then runs the auto-execution selector against the rank-1 trade if `TRADING_ENABLED=true`.

`ANTHROPIC_API_KEY` is required at boot (`mustEnv` in `internal/config/config.go`). When it's a local stub the picker is left nil and cron jobs short-circuit so the local Docker stack boots without making real API calls.

### Model version refresh policy

The picker model is configured via the `ANTHROPIC_MODEL` env var with the default defined as the `DefaultAnthropicModel` constant in `vibetradez.com/server/internal/config/config.go`.

**Any time work touches the trade picker or this default, fetch the official Anthropic Go SDK documentation and refresh the default to the current latest production model.** Anthropic publishes new model versions regularly; if the default sits stale, trade quality degrades silently. The page to read is:

- Anthropic Go SDK: <https://platform.claude.com/docs/en/api/sdks/go>

When updating, also bump the `ANTHROPIC_MODEL` default baked into `vibetradez.com/local/docker-compose.local.yml` so the local dev stack matches.

## CI/CD Pipeline

Triggered manually via GitHub Actions (`workflow_dispatch`). Runs on the production server via SSH. Merges to `main` no longer auto-deploy; trigger via the "Run workflow" button on the Actions tab or `gh workflow run main-pipeline.yml`. Linting is still split per-project so failures surface fast per side, but deploy and notify are each a single unified job.

```
            ┌──────────────────── LINTS (parallel × 4) ─────────────────────┐
            │  ┌──────────────┐                                              │
            ├─>│ Lint          │                                             │
            │  │ Portfolio     │                                             │
            │  │ (Biome)       │                                             │
            │  └──────────────┘                                              │
┌────────┐  │  ┌──────────────┐       ┌─────────────────────┐                │
│ Manual │─>│  │ Lint          │      │ Build                │                │
│dispatch│  │─>│ Trading FE    │─────>│ docker compose build │──┐             │
└────────┘  │  │ (Biome)       │      │ --no-cache (3 imgs)  │  │             │
 │  Sync    │  └──────────────┘       └─────────────────────┘  │             │
 │ git      │  ┌──────────────┐                                 │             │
 │ reset    │  │ Lint          │                                │             │
 │ --hard   │─>│ Trading BE    │                                │             │
 │          │  │ (golangci)    │                                │             │
 │          │  └──────────────┘                                 │             │
 │          │  ┌──────────────┐                                 │             │
 │          │  │ Lint          │                                │             │
 │          ├─>│ Auth Service  │                                │             │
 │          │  │ (golangci)    │                                │             │
 │          │  └──────────────┘                                 │             │
            └──────────────────────────────────────────────────────┼─────────┘
                                                                   │
                                                                   ▼
                                          ┌────────────────────────────────────┐
                                          │ Deploy                              │
                                          │ rollout jaycebordelon-com           │
                                          │ + rollout trading-frontend          │
                                          │ + force-recreate trading-server     │
                                          │ (each step: continue-on-error)      │
                                          └────┬──────────────────────┬─────────┘
                                               │                      │
                                               ▼                      ▼
                                      ┌──────────────────┐   ┌───────────────────┐
                                      │ Notify           │   │ Cleanup + Health  │
                                      │ One email, slate │   │ (only on full     │
                                      │ theme, per-site  │   │  success)         │
                                      │ status + /health │   │ prune + endpoints │
                                      └──────────────────┘   └───────────────────┘
```

Both Go lint jobs SSH into the same droplet. golangci-lint takes a process-wide flock at `$TMPDIR/golangci-lint.lock` (see [`pkg/commands/run.go`](https://github.com/golangci/golangci-lint/blob/master/pkg/commands/run.go) — `filepath.Join(os.TempDir(), "golangci-lint.lock")`); the lock is **not** inside `GOLANGCI_LINT_CACHE`, so disjoint cache dirs alone are not enough. Each job therefore exports its own `TMPDIR` (`$HOME/.tmp/golangci-lint-{trading,auth}-service`) and its own `GOLANGCI_LINT_CACHE` (`$HOME/.cache/golangci-lint-{trading,auth}-service`). With per-job lock files and per-job caches, the two Go lints run truly concurrently with each other and with the two Biome lints.

1. **Sync** — `git reset --hard origin/main`
2. **Lint / Portfolio** — Biome check on `jaycebordelon.com/` (runs in parallel with the other three lints)
3. **Lint / Trading Frontend** — Biome check on `vibetradez.com/client/` (runs in parallel with the other three lints)
4. **Lint / Trading Server** — `golangci-lint run ./...` on `vibetradez.com/server/`, with `TMPDIR=$HOME/.tmp/golangci-lint-trading-server` + `GOLANGCI_LINT_CACHE=$HOME/.cache/golangci-lint-trading-server` (runs in parallel with the other three lints)
5. **Lint / Auth Service** — `golangci-lint run ./...` on `auth.jaycebordelon.com/`, with `TMPDIR=$HOME/.tmp/golangci-lint-auth-service` + `GOLANGCI_LINT_CACHE=$HOME/.cache/golangci-lint-auth-service` (runs in parallel with the other three lints)
6. **Build** — Single `docker compose build --no-cache jaycebordelon-com trading-server trading-frontend` invocation, gated on all four lints passing. (`auth-service` is linted but not built/deployed by this pipeline — it ships via manual deploys.)
7. **Deploy** — One job with two sequential SSH steps (`continue-on-error: true` on each): (a) `docker rollout jaycebordelon-com`, (b) `docker rollout trading-frontend` + `docker compose up -d --force-recreate trading-server`. A final status step fails the job if either side failed, but both are always attempted. Per-site and overall statuses are exported as job outputs for the notify step.
8. **Notify** — One consolidated email with the slate/portfolio theme. Subject is `[PASSED|FAILED] jaycestuff - <short_sha>`. Body shows overall badge, per-site rows (jaycebordelon.com + vibetradez.com with individual PASSED/FAILED), commit metadata, and the trading-server `/health` table. Always fires unless the workflow is cancelled, so partial failures still produce an email.
9. **Cleanup** — `docker system prune -af` (volumes preserved so Traefik's cert storage survives). Runs only when deploy succeeded on both sides.
10. **Health Check** — Verify all endpoints + granular `/health` for trading server services (database, anthropic, schwab, api). The healthcheck step iterates `services | keys[]` so any new service added to the granular `/health` response is automatically gated without YAML changes. Runs only when deploy succeeded on both sides.
