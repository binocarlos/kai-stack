# CLAUDE.md

Conventions for this repo. Read before adding features so every session works the same way.

## Golden rule
**When you introduce a new convention, ask the user whether to record it here.** Keep entries terse and reference file paths instead of pasting code — this file is loaded into every session, so spend tokens sparingly.

## Stack
- **`api/`** — Go API: [Fiber](https://gofiber.io) HTTP, GORM over **Supabase** Postgres, **pgvector** search, an in-house migration system, and a simple Postgres-backed job queue (`SELECT ... FOR UPDATE SKIP LOCKED`, no external broker).
- **`frontend/`** — React + Vite SPA. All data goes through the Go API; the browser uses Supabase **only for auth** (Google OAuth via `supabase-js`), then sends the resulting JWT to the API as a Bearer token.

## Config & env
- Every env var is declared in `api/pkg/config/config.go` via `envconfig` tags — add new config there, nowhere else.
- `.env` is loaded in `api/main.go`; `.env.example` is the canonical reference.
- Supabase connection is `POSTGRES_*` (SSL required; pooler or direct host). `POSTGRES_AUTO_MIGRATE=true` applies migrations on boot.

## First-time setup (Supabase)
The Postgres connection is assumed to already work. To bring up **auth** (Google login), do this once, then `cp .env.example .env` and fill the values below.

**In the Supabase dashboard:**
1. **Authentication → Providers → Google**: enable it and paste a Google OAuth **Client ID + Secret**. (Create them in Google Cloud Console → Credentials → OAuth client ID → *Web application*, with authorized redirect URI `https://<project-ref>.supabase.co/auth/v1/callback`.)
2. **Authentication → URL Configuration**: set **Site URL** and add **Redirect URLs** for every origin you log in from — e.g. `http://localhost:8080` (dev) and your prod domain. The app redirects back to `window.location.origin`, which must be allow-listed here.
3. **Project Settings → API Keys**: copy a browser/client key — the new **publishable** key (`sb_publishable_…`, preferred) or the legacy **anon** key. (The RLS warning on it doesn't apply here: the browser uses Supabase only for auth, never its data API.) The **Project URL** is `https://<project-ref>.supabase.co`, where `<project-ref>` is the Project ID under **Settings → General**.
4. **JWT keys**: new projects use asymmetric signing keys by default — the API verifies tokens via the JWKS at `<SUPABASE_URL>/auth/v1/.well-known/jwks.json`, no secret needed. (Legacy HS256-only projects would need a code change; flagged, not supported yet.)

**Env vars to set** (see `.env.example` for the full annotated list):
- API (verifies Supabase JWTs): `SUPABASE_URL`, plus `AUTH_SUPABASE_ENABLED=true`, `SUPABASE_JWT_AUD=authenticated`, `AUTH_LOCAL_ENABLED` (keep `true` for the dev fixed-password fallback).
- Frontend (Vite inlines at build time; anon key is public): `VITE_SUPABASE_URL`, `VITE_SUPABASE_ANON_KEY`.
- Still required regardless: `POSTGRES_*`, `OPENAI_KEY`, `OPENAI_URL`, `SERVER_JWT_SECRET`, `SERVER_FIXED_PASSWORD`, `WORKER_SECRET`.

**Auth model**: the API verifies a Bearer JWT (Supabase via JWKS, or the local HS256 token) behind the `auth.Authenticator` seam (`api/pkg/auth/`), then maps the identity onto a row in the `profiles` table — that profile UUID is the app's canonical user id. Add a new provider by implementing `Authenticator`; nothing downstream changes.

## Server conventions (`api/pkg/server/`)
- Handlers are **plain methods on `*StackAPIServer`** that call the store/jobqueue directly. See `user.go` and `example_record.go`. **There is no generic resource/mapper/hook abstraction — write CRUD explicitly.**
- The server holds `store` (`*store.PostgresStore`) and `jobqueue` (`*jobqueue.Client`) as concrete structs.
- Auth: protect routes with the `apiServer.RequireAuth` middleware; read the user with `GetUserIDFromContext(c)` (see `auth.go`).
- Parse request bodies with `getRequestData[T](c)` (`server.go`).
- Error responses use the shape `{"error": "...", "code": "..."}`.

## Adding a new record type (full ORM + CRUD)
Use `example_record` as the end-to-end template:
1. **Model**: add the struct to `api/pkg/types/models.go` with `gorm` + `json` tags.
2. **Migration**: add a table migration (see below).
3. **Store**: add `api/pkg/store/<name>.go` with a repo embedding `*Repository[T]` (free `Create`/`FindByID`/`FindAll`/`Update`/`Delete`) plus any hand-written queries; expose it from `PostgresStore` with an accessor (e.g. `ExampleRecords()`).
4. **Handlers**: copy `api/pkg/server/example_record.go` to `<name>.go`, adapt the CRUD methods, and add `Register<Name>Routes()`; call it from `NewServer` in `server.go`.
5. **Frontend types**: mirror the struct in `frontend/src/types/gotypes.ts` (kept in sync by hand).

## Adding a DB migration (`api/pkg/store/migrations/`)
- Create `00000N_name.go` with a func `func(ctx context.Context, m Migrator) error` that calls `m.ExecSQL(...)`.
- **Append** it to `All` in `migrations/all.go` (`{Name: "00000N_name", Up: Name}`).
- Append-only: never rename, reorder, or edit an applied migration — state is tracked in the `go_schema_migrations` ledger and each migration runs in its own transaction.

## Background jobs (`api/pkg/jobqueue/`)
Template: `example.go`.
1. Define a payload struct and `const <Name>JobKind = "..."`.
2. Write `register<Name>Handler(c *Client)` calling `c.Register(<Name>JobKind, func(ctx, payload) error { ... })`.
3. Register it in `NewClient` (`client.go`).
4. Enqueue from a handler: `apiServer.jobqueue.Enqueue(context.Background(), <Name>JobKind, payload)` — use a fresh context (not the request's) and don't fail the request on enqueue error. The worker (`worker.go`) polls and runs handlers automatically.

## Vector search
`ExampleRecordRepository.UpsertEmbedding` / `SearchSimilar` (`api/pkg/store/example_record.go`) show the pattern: pgvector type binding, cosine distance `<=>`, HNSW index from migration `000002`, 1536-dim embeddings.

## Frontend (`frontend/`)
- Stack: Vite + React + router5 + TanStack Query + axios + MUI.
- **Add a page**: component in `src/pages/`, then a route in `src/routes.tsx` (`name`/`path`/`meta`/`render`, optional `processRoute` auth guard).
- **Data**: hooks in `src/hooks/` using `useQuery`/`useMutation` + axios against `API_BASE_URL` (`src/constants/system.ts`, = `/api/v1`). The Bearer token is attached to every request by an axios interceptor in `src/supabase.ts` (Supabase session token, else the local token); login state lives in `src/contexts/account.tsx`. Surface errors with `extractErrorMessage` (`src/utils/apitools.ts`) + `useSnackbar`.

## Run / build
- API: `cd api && go build ./... && go vet ./...`; run the server via the cobra `serve` command.
- Frontend: `cd frontend && npm run dev` (port 8080) / `npm run build`.
