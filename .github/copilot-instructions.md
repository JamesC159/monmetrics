# MonMetrics AI Coding Agent Instructions

## Project Overview

MonMetrics is a **trading card price analysis platform** combining a **Go backend** (pure stdlib, no frameworks) with a **React 19 frontend** (Vite + TypeScript). The architecture emphasizes minimalism, type safety, and OWASP security compliance.

## Architecture & Key Decisions

### Backend: Zero-Framework Go Pattern

- **Pure net/http stdlib** - No Gin, Echo, or Chi. All routing via `http.NewServeMux()` with Go 1.22+ path patterns
- **Middleware chaining** - Custom `middleware.Chain()` applies CORS, security headers, rate limiting, and auth in order
- **Two-tier routing** - Public API (`/api/*`) and protected API (`/api/protected/*`) with separate middleware stacks
- **MongoDB with native driver** - Collections: `cards`, `prices`, `listings`, `market_data`, `saved_charts`, `users`
- **Manual JWT** - HMAC-SHA256 signing without external libs. See `handlers.generateJWT()` and `middleware.validateJWT()`

**Critical**: When adding endpoints, mount them on the correct mux (`apiMux` for public, `protectedMux` for auth-required) in `cmd/server/main.go`.

### Frontend: React 19 SSR with Vite

- **Native SSR** - Uses React 19's built-in server rendering (no Next.js/Remix)
- **Entry points**: `entry-client.tsx` (browser), `entry-server.tsx` (SSR), `server.js` (Node HTTP server)
- **State management** - React Context for global state (AuthContext, ToastContext). No Redux/Zustand.
- **API client pattern** - Singleton `apiClient` (in `utils/api.ts`) handles all HTTP, auto-adds Bearer tokens from localStorage

**Critical**: Always use `apiClient` methods. Never `fetch()` directly. Token management is automatic.

## Development Workflow

### Essential Commands (Use Makefile)

```bash
make full-setup   # One-time: installs deps, creates .env, starts MongoDB, seeds DB
make dev          # Start both servers (backend :8080, frontend :3000)
make seed         # Re-populate DB with sample data (11 cards, 5yr price history)
make reset        # Nuclear option: drops DB, cleans builds, requires full-setup after
```

**Important**: Backend requires Go 1.24.2 in `/usr/local/go`. If Go version mismatch errors occur, run:

```bash
export PATH=/usr/local/go/bin:$PATH
cd backend && go clean -cache -modcache && go mod tidy
```

### Build & Test Patterns

- **Backend build**: `go build -o bin/server cmd/server/main.go` (produces single binary)
- **Frontend build**: `npm run build` (creates `dist/` with SSR-ready static assets)
- **Tests**: `make test-backend` (Go tests), `make test-frontend` (Vitest)
- **No hot reload for Go** - Must restart `make dev` after backend code changes

## Security & Validation Patterns

### OWASP-Compliant Input Handling

All user inputs go through this flow (see `handlers.validateRegisterRequest()`):

1. **Normalize** - `normalizeEmail()` lowercases/trims, `sanitizeName()` removes control chars
2. **Validate** - Regex for email (RFC 5322), password strength (8+ chars, upper/lower/digit/special), name length (2-50)
3. **Hash** - bcrypt cost 12 (~400ms/hash) for passwords. Never store plaintext.

**Critical**: Frontend validation is UX only. Backend MUST re-validate everything. Never trust client data.

### Password Requirements (enforced in `validatePassword()`)

- Min 8, max 128 chars
- At least 1 uppercase, 1 lowercase, 1 digit, 1 special character
- Bcrypt hashed with cost 12 (production-ready balance of security vs performance)

### Rate Limiting

Default: 100 requests/60s per IP. Token bucket algorithm in `middleware.RateLimit()`. Adjust via `RATE_LIMIT_REQUESTS`/`RATE_LIMIT_WINDOW` env vars.

## Data Flow & Integration Points

### Search Architecture (Cards)

MongoDB regex search + term matching (see `buildSearchFilter()`):

- Full-text search disabled due to unreliability. Uses case-insensitive regex on `name`, `set`, `game` fields
- `search_terms` array (lowercase tokens) enables partial matching
- Fallback logic: If complex query fails, retry with simpler filter
- Returns empty array `[]` not null for zero results (frontend expects array)

**Critical**: When adding searchable fields to cards, update both regex patterns AND `search_terms` array generation in seeder.

### Price History Flow

1. Frontend requests `/api/cards/{id}/prices?range=30d`
2. Backend queries 3 collections: `prices` (time-series points), `listings` (current market), `market_data` (OHLC aggregates)
3. All queries filter by `card_id` + timestamp range
4. Response shape: `{ prices: [], listings: [], market_data: [], card_id, time_range }`

**Time ranges**: `1d`, `7d`, `30d`, `90d`, `1y`, `5y` (default: `30d`)

### Authentication Flow

1. **Register/Login** → Returns JWT + user object
2. **Frontend** stores token in localStorage, sets on `apiClient`
3. **Protected requests** → `apiClient` auto-adds `Authorization: Bearer <token>` header
4. **Middleware** validates JWT signature, checks expiry, adds `claims` to request context
5. **Handlers** read user ID from `r.Context().Value("claims")`

**Critical**: JWT exp is 24h. No refresh token mechanism yet. Consider adding if needed.

## TypeScript & Go Type Alignment

### Field Name Conventions

- **Go struct tags**: Use `json:"camelCase"` for JSON, `bson:"snake_case"` for MongoDB
- **TypeScript**: Matches Go JSON tags exactly (camelCase)
- **Example**: Go `FirstName string json:"firstName"` ↔ TS `firstName: string`

**Common mistake**: Using snake_case in JSON tags breaks frontend. Always camelCase for API contracts.

### MongoDB ID Handling

- **Go**: `primitive.ObjectID` with BSON tag `_id,omitempty`
- **JSON**: Serializes as `id` (set in struct tag: `json:"id"`)
- **Frontend**: Use `.hex()` for string representation when needed

## Component Patterns (Frontend)

### Context Usage

- **AuthContext**: Provides `user`, `login()`, `register()`, `logout()`, `isAuthenticated`
- **ToastContext**: Provides `showToast(message, type)` for notifications
- Always wrap app with `<AuthProvider>` and `<ToastProvider>` in `App.tsx`

### Error Handling Pattern

```typescript
try {
  const result = await apiClient.someMethod()
} catch (error) {
  if (error instanceof Error) {
    showToast(error.message, 'error')
  }
}
```

API errors include `.status` and `.details` properties. Use for specific handling.

### Route Protection

Use conditional rendering based on `isAuthenticated`:

```tsx
{
  isAuthenticated ? <Dashboard /> : <Navigate to='/login' />
}
```

Backend double-checks auth via middleware. Frontend guards are UX only.

## Common Pitfalls & Solutions

### "MongoDB connection refused"

```bash
docker ps  # Check monmetrics_mongo is running
make reset && make setup  # Restart MongoDB container
```

### "Go version mismatch" (compile error)

```bash
export PATH=/usr/local/go/bin:$PATH
cd backend && go clean -cache -modcache
```

### "Search returns no results" (when data exists)

Check `search_terms` array population in seeder. Must be lowercase tokens matching query terms.

### "CORS error in browser"

Ensure `CORS_ORIGINS=http://localhost:3000` in `backend/.env`. Middleware allows localhost variations in dev mode.

## Code Style Conventions

### Go

- **No blank identifiers** for errors - handle or log explicitly
- **Context timeouts**: Always use `context.WithTimeout()` for DB ops (5-15s)
- **Printf statements**: Keep for debugging, no formal logger (stdlib only)
- **Error responses**: Use `h.sendError(w, msg, status, details)` helper for consistency

### TypeScript

- **Type imports**: Use `import type { ... }` for type-only imports
- **Optional chaining**: Use `?.` for nested objects that might be undefined
- **Array initialization**: Prefer `[]` over `null` for empty arrays (backend returns `[]`)
- **Env vars**: Always use `getEnvVar()` helper with fallbacks for SSR safety

## Future Enhancements (TODOs)

### Priority 1

- Email verification on registration (add `email_verified` boolean to User model)
- Token refresh mechanism (current JWT expires in 24h, no renewal)
- Password reset flow (requires email integration)

### Priority 2

- CAPTCHA on registration (recommend hCaptcha for free tier)
- Advanced technical indicators (Bollinger Bands, RSI, MACD) in price charts
- WebSocket for real-time price updates (consider using gorilla/websocket despite stdlib rule)

### Priority 3

- Elasticsearch for search (current regex approach doesn't scale beyond 10k cards)
- Redis for session caching (reduce MongoDB load on repeated dashboard fetches)
- Mobile app (React Native - can reuse `apiClient` logic)

## Debugging & Troubleshooting

### Backend Logs

Watch Go server stdout for:

- `🏥 Health check` - Shows DB ping status
- `📊 Database Status` - Connection info on startup
- HTTP logs include: method, path, status, duration, client IP

### Frontend DevTools

- Network tab: Check `Authorization` header presence on protected routes
- Console: `apiClient` errors include full response body
- localStorage: Key `authToken` stores JWT (decode at jwt.io for debugging)

### MongoDB Inspection

```bash
docker exec -it monmetrics_mongo mongosh
use monmetrics
db.cards.countDocuments()  # Should return 11 after seeding
db.users.find()  # Check registered users
```

## Testing Strategy

### Backend (Go)

- Unit tests for: password hashing/verification, email validation, JWT generation
- Integration tests for: search filters, price queries, auth flow
- Run: `cd backend && go test ./...`

### Frontend (Vitest)

- Component tests for: form validation, auth flow, error states
- API client mocks for isolated testing
- Run: `cd frontend && npm test`

**Critical**: Test bcrypt hashing separately (slow). Mock in integration tests to avoid 400ms overhead per case.

---

## Quick Reference

| Task              | Command                                               | Notes                             |
| ----------------- | ----------------------------------------------------- | --------------------------------- |
| Start dev         | `make dev`                                            | Backend :8080, Frontend :3000     |
| Add endpoint      | Edit `cmd/server/main.go`                             | Choose `apiMux` or `protectedMux` |
| New model         | Edit `internal/models/models.go`                      | Sync Go tags with TS types        |
| New frontend page | Add to `pages/`, register route in `App.tsx`          | Use AuthContext for protection    |
| DB schema change  | No migrations - MongoDB is schemaless                 | Update models + re-seed           |
| Deploy            | `make build`, run `bin/server`, serve `frontend/dist` | Set production env vars           |

**First-time setup**: `make full-setup && make dev` → Visit http://localhost:3000

---

**Last Updated**: November 2025 | **Go**: 1.24.2 | **Node**: 18+ | **MongoDB**: 7.0
