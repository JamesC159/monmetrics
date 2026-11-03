# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd prime` at session start for the full workflow; the managed Beads sections near the end of this file cover the basics.

## Beads in This Repo

- Issue IDs use the `monmetrics-` prefix. Issues live in embedded Dolt under `.beads/embeddeddolt/`. Cross-machine sync uses `bd dolt push/pull` to `refs/dolt/data` on this repo's GitHub origin, separate from the code branches. `.beads/issues.jsonl` is a passive export: don't treat it as the source of truth, and don't `bd import` during normal work.
- The active profile is **conservative** (`bd config get agent.profile`). Don't commit, push, or run `bd dolt push` unless the user asks.
- Before writing code, create or claim the bead. When work turns up follow-ups, file them with `bd create "..." --deps discovered-from:<current-id>` instead of leaving TODOs.
- **Usage metrics stay off.** bd ships with anonymous usage metrics on (sent to `gastownhall-eventsapi.com`). On this machine they're disabled in `~/.config/bd/config.yaml`. In any other environment (new machine, container, CI), run `bd metrics off` or export `BD_DISABLE_METRICS=1` before other bd commands. Claude Code sessions get that variable from `.claude/settings.json`. `bd metrics` must report OFF. Never run `bd metrics on`.

<!-- Project context below is shared: keep it identical in AGENTS.md and CLAUDE.md. -->

## Project Overview

MonMetrics is a trading-card price-analysis web app. Users can:

- search and browse a catalog of singles and sealed products
- read price history on interactive charts with indicators and drawing tools
- save charts and set price alerts
- track a portfolio with valuations
- draft marketplace listings

Catalog data and daily prices come from TCGCSV, a TCGplayer mirror. An eBay sold-listing scraper exists but is blocked (HTTP 403), so the features that need it sit behind feature flags that are off by default. `ENABLED_GAMES` defaults to `Magic`, so only Magic: The Gathering is shown unless that setting changes.

## Build & Test

Go 1.24.2 lives in `/usr/local/go/bin`. The Makefile adds it to `PATH`; a bare shell may need `export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH`.

```bash
make full-setup   # EMPTY database only: deps, .env files, MongoDB container, then sample seed data (see "Protect real data")
make setup        # safe on any database: deps, missing .env files, MongoDB container; never seeds
make dev          # MongoDB (docker-compose) + backend :8080 (air hot reload) + Vite :3000

# Backend quality gate (needs no MongoDB or network)
cd backend && go build ./... && go vet ./... && go test ./...

# Frontend quality gate
cd frontend && npm run type-check && npm test
cd frontend && npm run build   # client bundle + SSR bundle (dist/server)
```

- Run the gate for every area you touched before closing a bead.
- `npm run lint` fails today because the repo has no ESLint config. That isn't a regression.
- Data jobs call live third-party services. Run them only when the task needs them:
  - `make catalog-sync ARGS="--group 3170 --dry-run"` imports the TCGCSV catalog. `make catalog-demo` ingests one set end to end.
  - `make prices ARGS="-max-groups 3 -dry-run"` fetches daily TCGplayer prices. Run it at most once a day: TCGCSV bans clients over 10k requests per 24h, and price files must not be re-fetched within 24h.
  - `make scrape` pulls eBay sold listings and currently gets 403s. `go run ./cmd/scraper -fixture <page.html>` parses a saved page without touching eBay.
  - `make seed` replaces the catalog with sample cards, 5 years of synthetic price history and a test user. Read the rule below first.

**Protect real data.** Never run `make full-setup` or `make seed` on a database where `make catalog-sync`, `make catalog-demo`, `make prices` or `make scrape` has ever run. The seeder deletes everything in `cards`, `prices`, `listings` and `featured_content` and refills them with sample data. It refuses to run (exit 1) when it finds anything those jobs leave behind; see `realDataMarks` in `backend/cmd/seeder/guard.go`. A new job that writes real data must add its trace there.

- Never set `SEED_FORCE=true`; it bypasses that guard.
- Never run `make reset` unless the user asks; it deletes the MongoDB volume outright.
- If you can't tell whether the database holds real data, assume it does. Use `make setup` to get dependencies and the container running.

## Architecture Overview

The monorepo holds a stdlib Go API over MongoDB and a React single-page app with an SSR build.

```text
backend/                         module github.com/jamesc159/monmetrics (net/http + mongo-driver v1, no web framework)
  cmd/server/main.go             every route and middleware stack; starts the alert/scraper/catalog runners when env enables them
  cmd/{catalog,prices,scraper,seeder}   one-shot CLIs behind the make targets above
  configs/                       Config + Load(): env vars with defaults; .env only fills variables that are unset
  internal/handlers/             HTTP handlers as methods on *Handlers, one file per domain
  internal/middleware/           Chain, CORS, SecurityHeaders, RateLimit, RequestLogger, AuthRequired (hand-rolled HS256 JWT), AdminRequired
  internal/models/               Mongo documents and API request/response structs, one file per domain
  internal/database/             Mongo connection; every index is declared in createIndexes
  internal/catalog/              TCGCSV category -> group -> product ingest, checkpoints, NormalizeProduct, BuildSearchTerms
  internal/pricing/              daily price snapshots -> prices (one row per card/printing/day; a rerun overwrites that day)
  internal/scraper/              eBay sold-listing source, relevance and grade classification, tiered scheduler, daily aggregates
  internal/alerts/               alert rules (thresholds, EMA) and the background Evaluator -> notifications + email
  internal/features/             Mongo-backed feature flags with declared defaults and "blocker" explanations
  internal/valuation/            portfolio valuation and grade multipliers
  internal/marketplace/          provider registry (only "mock" exists), comps, listing prefill
  internal/notify/, secure/      SMTP sender (logs instead when SMTP_HOST is empty); AES-GCM for linked-account tokens at rest
frontend/                        React 19 + TypeScript, Vite 8, Tailwind 3, react-router-dom v7, lightweight-charts v5
  src/App.tsx                    providers and the route table
  src/pages/                     route components
  src/components/                by area: chart/ (PriceChart, indicators, drawing tools), dashboard/, catalog/, card/, portfolio/, marketplace/, alerts/
  src/context/                   Auth, Toast, Favorites, Features providers (no Redux)
  src/utils/api.ts               the ApiClient singleton, the only HTTP layer
  src/types/index.ts             TypeScript mirrors of the backend JSON
  entry-client.tsx, entry-server.tsx, server.js   hydration entry, SSR render, Express SSR server for production
```

**Request path.** `cmd/server/main.go` mounts three muxes:

- `/api/*` is public and runs CORS, SecurityHeaders, RateLimit and RequestLogger.
- `/api/protected/*` adds `AuthRequired`.
- `/api/admin/*` adds `AdminRequired`, an allowlist from `ADMIN_EMAILS`.

`/health` has no middleware.

**Data flow.**

- Catalog ingest upserts `cards`, which are unique on `catalog_source` + `external_id`.
- Price ingest writes daily rows to `prices` (source `tcgplayer`, one series per printing).
- The scraper writes `sold_listings` and `price_summaries`.
- The alert evaluator reads `prices` and writes `notifications`, plus email.

User data lives in `users`, `saved_charts`, `price_alerts`, `favorites`, `portfolio_items`, `linked_accounts` and `marketplace_listings`.

**Feature flags** live in `internal/features`: `ebay.sold_listings`, `ebay.grades`, `price.summary_windows`, `marketplace.comps` and `catalog.all_games`. All default to off, and admins toggle them at `/admin`. Handlers check `h.features.Enabled(key)` and the UI reads `FeaturesContext`. Put any surface whose data source isn't live yet behind a flag the same way.

## Conventions & Patterns

**Backend (Go)**

- HTTP uses the standard library only. Routes are Go 1.22 method+path patterns (`"GET /cards/{id}"`), read with `r.PathValue("id")`. Don't add a router or web framework.
- A new endpoint needs four pieces:
  - a handler in the matching `internal/handlers/<domain>.go`
  - registration on the right mux (public, protected or admin) in `cmd/server/main.go`
  - an `apiClient` method in the frontend
  - a TypeScript type in the frontend
- Use the existing helpers: `decodeJSON` (1 MB cap, `DisallowUnknownFields`), `sendJSON`, `h.sendError` and `userIDFromRequest`.
- JSON tags are **snake_case and match the bson tags** (`bson:"current_price" json:"current_price"`). `frontend/src/types/index.ts` mirrors them field for field. `.github/copilot-instructions.md` says camelCase, which is out of date.
- Wrap every Mongo call in `context.WithTimeout`. Declare indexes for new query shapes in `database.createIndexes`. Return `[]`, never `null`, for empty lists.
- Treat input as data:
  - run `regexp.QuoteMeta` before building a regex from input
  - whitelist sort and filter values (see `searchSorts` and `validGradePattern`)
  - validate on the server even when the UI already did
- Catalog-facing queries go through `h.applyEnabledGames(filter)`, so `ENABLED_GAMES` and `catalog.all_games` apply.
- New config gets a field on `configs.Config`, parsing in `Load()` with a default and a fallback for bad values, and an entry in `backend/.env.example`.
- External systems sit behind interfaces: `scraper.SoldSource`, the catalog source, `marketplace.Provider` and `notify.Sender`. Tests use fakes or fixtures and never touch the network or MongoDB. They're `_test.go` files next to the code and use plain `testing`.
- `alerts.EMA` must match `calculateEMA` in `frontend/src/utils/indicators.ts`; `TestEMAMatchesFrontend` checks this.
- Logging is `fmt`/`log` with emoji prefixes. Match the file you're in.
- Respect third-party limits. TCGCSV asks for about 100 ms between calls (`CATALOG_REQUEST_DELAY`, `PRICE_REQUEST_DELAY`). The scraper has its own delay (`SCRAPER_REQUEST_DELAY`) and a daily budget.

**Frontend (React/TypeScript)**

- All HTTP goes through `apiClient` in `src/utils/api.ts`, which adds the Bearer token from `localStorage.authToken`. Don't call `fetch` directly.
- Global state is React Context, with the providers set up in `App.tsx`. Custom hooks live in `src/hooks/`.
- Import through the `@/` alias, and use `import type` for type-only imports.
- Code style is single quotes, no semicolons, 2-space indent, and single-quoted JSX attributes.
- Style with Tailwind utilities from the graphite palette in `tailwind.config.ts` and `src/theme`, and combine classes with `cn()`. `THEME_GUIDE.md` describes an older gold/cyan palette.
- Code must be SSR-safe. Guard `window` and `localStorage` with `typeof window !== 'undefined'`, and read env vars with a fallback as `getEnvVar` does.
- Vitest tests sit next to the code as `*.test.ts`.

**Repo**

- `.gitignore` ignores `*.md` apart from the agent files re-included at the end of it. A new tracked doc needs its own negation rule or `git add -f`.
- `README.md`, `.github/copilot-instructions.md` and the root `*_SUMMARY.md` and analysis docs are partly stale. Trust the code and this file over them.
- Never commit `backend/.env` or `frontend/.env.local`. `TOKEN_ENCRYPTION_KEY` is required when `ENVIRONMENT=production`.

<!-- End of shared project context. -->

## Non-Interactive Shell Commands

**ALWAYS use non-interactive flags** with file operations to avoid hanging on confirmation prompts.

Shell commands like `cp`, `mv`, and `rm` may be aliased to include `-i` (interactive) mode on some systems, causing the agent to hang indefinitely waiting for y/n input.

**Use these forms instead:**
```bash
# Force overwrite without prompting
cp -f source dest           # NOT: cp source dest
mv -f source dest           # NOT: mv source dest
rm -f file                  # NOT: rm file

# For recursive operations
rm -rf directory            # NOT: rm -r directory
cp -rf source dest          # NOT: cp -r source dest
```

**Other commands that may prompt:**
- `scp` - use `-o BatchMode=yes` for non-interactive
- `ssh` - use `-o BatchMode=yes` to fail instead of prompting
- `apt-get` - use `-y` flag
- `brew` - use `HOMEBREW_NO_AUTO_UPDATE=1` env var

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:46cd31e7 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/core-concepts/sync-concepts.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   bd dolt push
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/core-concepts/sync-concepts.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->
