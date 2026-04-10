# Personal Monorepo

Jayce Bordelon's production monorepo. All services are deployed to a single Digital Ocean droplet running Docker Compose behind Traefik as a reverse proxy with automatic Let's Encrypt TLS.

## Architecture

**Single-server monolithic deployment.** Traefik routes incoming HTTPS requests to the correct container by hostname and path:

- `jaycebordelon.com` / `www.jaycebordelon.com` → Next.js portfolio (port 3000)
- `jaycetrades.com` → `/api/*`, `/auth/*`, `/admin/*`, `/health` → Go API server (port 8080, priority 20)
- `jaycetrades.com` → everything else → Next.js trading frontend (port 3001, priority 10)

## Project Structure

```
personal-monorepo/
├── jaycebordelon.com/           # Personal portfolio & blog
│   ├── app/                     # Next.js 16 App Router pages
│   ├── components/              # React components + shadcn/ui
│   ├── content/                 # MDX blog posts
│   ├── lib/                     # Utilities
│   ├── Dockerfile               # Multi-stage Node.js build
│   └── package.json             # Next.js 16, React 19, Tailwind v4
│
├── jaycetrades.com/
│   ├── server/                  # Go API server (trading backend)
│   │   ├── cmd/scanner/         # Main entry point, cron jobs, workflows
│   │   ├── internal/
│   │   │   ├── config/          # Environment variable loading
│   │   │   ├── email/           # Resend email client
│   │   │   ├── schwab/          # Schwab OAuth + Market Data API
│   │   │   ├── sentiment/       # Reddit WSB sentiment scraper
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
│   ├── main-pipeline.yml       # Orchestrator: sync → lint → build → deploy → cleanup → healthcheck → notify
│   ├── sync.yml                # Git pull on production server
│   ├── lint.yml                # Biome (JS/TS) + golangci-lint (Go)
│   ├── build.yml               # Docker compose build all services
│   ├── deploy.yml              # Rolling deployment via docker rollout
│   ├── cleanup.yml             # Post-deploy docker system prune
│   ├── healthcheck.yml         # Endpoint verification + granular /health
│   ├── notify.yml              # Email notification with pipeline results
│   └── cd.yml                  # Standalone manual trigger pipeline
│
├── docker-compose.yml          # All services + Traefik config
├── .env                        # Secrets (gitignored)
└── CLAUDE.md                   # This file
```

## Tech Stack

| Project | Stack |
|---------|-------|
| jaycebordelon.com | Next.js 16, React 19, Tailwind CSS v4, shadcn/ui (new-york), MDX, Framer Motion |
| jaycetrades.com/client | Next.js 16, React 19, Tailwind CSS v4, shadcn/ui (new-york), Recharts v3, TradingView Lightweight Charts |
| jaycetrades.com/server | Go 1.23, PostgreSQL (Digital Ocean managed), OpenAI GPT-5.4, Schwab Market Data API, Resend email |
| Infrastructure | Docker Compose, Traefik v2.10, Let's Encrypt, Digital Ocean Droplet |

## Database

PostgreSQL hosted on Digital Ocean Managed Databases. Connection string is in `.env` as `DATABASE_URL`. The Go server auto-migrates schema on startup (CREATE TABLE IF NOT EXISTS).

## Key Environment Variables (.env)

- `DATABASE_URL` — PostgreSQL connection string (required, no default)
- `RESEND_API_KEY` — Email delivery
- `OPENAI_API_KEY` — Trade analysis LLM
- `SCHWAB_APP_KEY` / `SCHWAB_SECRET` — Market data OAuth
- `ADMIN_KEY` — Protects `/admin/announce` broadcast endpoint
- `EMAIL_RECIPIENTS` — Seed subscribers on first boot

## Development Rules

### Always lint before pushing

Run these checks before every push. CI will fail if they don't pass:

```bash
# Go
cd jaycetrades.com/server && gofmt -w . && go vet ./...

# Next.js (both projects)
cd jaycebordelon.com && npx biome check .
cd jaycetrades.com/client && npx biome check .
```

### Always read the latest documentation

When working with Next.js, shadcn/ui, Tailwind CSS, Recharts, or any external library, **always fetch and read the current documentation** before writing code. Do not rely on recalled syntax or API signatures — they may be outdated. This applies even if it takes extra time. Incorrect assumptions about APIs cause more rework than the time saved by skipping docs.

### Always use feature branches

Never push directly to `main`. Create a descriptive branch, push there, and let the user handle PRs and merging.

### Design system consistency

Both Next.js frontends share the same design tokens (CSS variables in `globals.css`), font stack (Plus Jakarta Sans, JetBrains Mono), and shadcn/ui configuration (new-york style, neutral base color, lucide icons). Any UI changes must be consistent across both sites.

## API Protection

All `/api/*` routes on the trading server require the `X-JT-Source` header. Without it, requests return 403. The Next.js frontend includes this header on every fetch call. The `/admin/announce` endpoint requires `X-Admin-Key` header matching the `ADMIN_KEY` env var.

## Trading Server Workflows

The Go server runs three cron jobs in Eastern Time:
- **9:25 AM Mon-Fri** — Scrape Reddit sentiment, call OpenAI for 10 ranked trade picks, save to DB, email subscribers
- **4:05 PM Mon-Fri** — Fetch closing prices from Schwab, compute EOD P&L, save summaries, email subscribers
- **4:30 PM Fridays** — Aggregate weekly performance, compute stats (win rate, Sharpe, drawdown), email weekly report

Market holidays are hardcoded in `cmd/scanner/main.go`. Jobs skip on holidays and weekends.

## Common Operations

### Send announcement to all subscribers
```bash
curl -X POST https://jaycetrades.com/admin/announce \
  -H "X-Admin-Key: <ADMIN_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"subject": "...", "badge": "...", "headline": "...", "sections": [{"title": "...", "body": "..."}], "cta_text": "...", "cta_url": "..."}'
```

### Re-authorize Schwab OAuth
Visit `https://jaycetrades.com/auth/schwab` in a browser. Tokens are stored in the `oauth_tokens` table and auto-refresh.

### Check server health
```bash
curl https://jaycetrades.com/health | jq
```
Returns per-service status for database, LLM (OpenAI), Schwab, and API with latencies.

### Docker commands on production
```bash
ssh jayce@<server>
cd ~/personal-monorepo
docker compose logs trading-server --tail 50    # View Go server logs
docker compose logs trading-frontend --tail 50  # View Next.js logs
docker compose restart trading-server           # Restart Go server
docker compose up -d --force-recreate trading-server  # Full recreate
```

## Planned Migration

The trade analysis LLM is currently OpenAI GPT-5.4 (strong at sentiment analysis). A migration to Anthropic Claude is planned. This will require changes in `server/internal/trades/analyzer.go` (API client), `server/internal/trades/prompt.go` (prompt format), and the health check model probe in `server/internal/server/server.go`.

## CI/CD Pipeline

Triggered on push to `main` or manual dispatch. Runs on the production server via SSH:

1. **Sync** — `git reset --hard origin/main`
2. **Lint** — Biome + gofmt + go vet
3. **Build** — `docker compose build --no-cache` all services
4. **Deploy** — `docker rollout` for web apps, `docker compose up -d --force-recreate` for background services
5. **Cleanup** — `docker system prune -af --volumes` to reclaim disk space
6. **Health Check** — Verify all endpoints + granular `/health` for trading server services (database, LLM, Schwab, API)
7. **Notify** — Email with full pipeline status, commit info, and health results
