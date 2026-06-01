# CLAUDE.md

Conventions for this repo. Read before adding features so every session works the same way.

## Golden rule
**When you introduce a new convention, ask the user whether to record it here.** Keep entries terse and reference file paths instead of pasting code — this file is loaded into every session, so spend tokens sparingly.

## Stack
- **`api/`** — Go API: [Fiber](https://gofiber.io) HTTP, GORM over **Supabase** Postgres, **pgvector** search, an in-house migration system, and a simple Postgres-backed job queue (`SELECT ... FOR UPDATE SKIP LOCKED`, no external broker).
- **`frontend/`** — React + Vite SPA. All data goes through the Go API (no Supabase client in the browser).

## Config & env
- Every env var is declared in `api/pkg/config/config.go` via `envconfig` tags — add new config there, nowhere else.
- `.env` is loaded in `api/main.go`; `.env.example` is the canonical reference.
- Supabase connection is `POSTGRES_*` (SSL required; pooler or direct host). `POSTGRES_AUTO_MIGRATE=true` applies migrations on boot.

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
- **Data**: hooks in `src/hooks/` using `useQuery`/`useMutation` + axios against `API_BASE_URL` (`src/constants/system.ts`, = `/api/v1`). The Bearer token is set globally in `src/contexts/account.tsx`. Surface errors with `extractErrorMessage` (`src/utils/apitools.ts`) + `useSnackbar`.

## Run / build
- API: `cd api && go build ./... && go vet ./...`; run the server via the cobra `serve` command.
- Frontend: `cd frontend && npm run dev` (port 8080) / `npm run build`.
