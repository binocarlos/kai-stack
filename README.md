# kai-stack

A blueprint monorepo: a Go API, a React frontend, and a self-contained **Solana
reference app**. Conventions for working in the codebase live in
[`CLAUDE.md`](CLAUDE.md).

## Layout

| Path | What it is |
|------|------------|
| [`api/`](api/) | Go API — Fiber HTTP, GORM over Supabase Postgres, pgvector search, a Postgres-backed job queue. |
| [`frontend/`](frontend/) | React + Vite SPA. Talks to the Go API; uses Supabase only for auth. |
| [`solana/`](solana/) | Self-contained Solana app: Anchor program → typed `@solana/kit` client → React demo. **Docker-only, isolated from the rest.** |

## Running the main stack (API + frontend)

```bash
cp .env.example .env        # fill in the values (see CLAUDE.md "First-time setup")
./stack start               # brings up the API + frontend via Docker (see docker-compose*.yml for ports)
./stack stop
```

Build/typecheck directly: `cd api && go build ./...` · `cd frontend && npm run dev`.

## Running the Solana stack

The Solana app lives in [`solana/`](solana/) and is driven by its own
`./stack` script. You need **only Docker** — the Rust/Solana/Anchor/Node
toolchain all runs in a container. From the `solana/` directory:

```bash
cd solana
cp .env.example .env            # defaults work for local dev

./stack build                   # compile the Anchor program (emits the IDL)
./stack codegen                 # generate the @solana/kit client from the IDL
./stack test                    # integration tests (kit client vs a validator)
./stack test-unit               # fast in-process tests (LiteSVM)

# Run it for real, locally:
./stack localnet                # terminal 1: a local validator (surfpool)
./stack deploy-localnet         # terminal 2: deploy the program to it
./stack app                     # the React demo at http://localhost:5173 — connect a wallet

# Deploy to a public cluster:
./stack deploy-devnet           # also: deploy-testnet / deploy-mainnet
```

Run `./stack` with no arguments to list every command. Full details:
- [`solana/README.md`](solana/README.md) — overview and the program → types → frontend pipeline
- [`solana/docs/development.md`](solana/docs/development.md) — the local dev loop
- [`solana/docs/deployment.md`](solana/docs/deployment.md) — devnet/mainnet + verifiable builds
- [`solana/docs/integration.md`](solana/docs/integration.md) — folding it into the main stack

> Note: `solana/program/counter-keypair.json` is a committed throwaway dev
> program id so the repo clones-and-runs — generate a fresh one for any real
> deployment (see the deployment doc).
