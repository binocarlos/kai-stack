# Solana reference app

A self-contained, Docker-first blueprint for the **whole Solana loop**: write a
program → generate typed TypeScript from it → call it from a React frontend (and
from plain TS) → deploy to devnet/mainnet.

It's designed so that **someone who has never written Solana code can follow it
end to end.** You need only **Docker** — the Rust/Solana/Anchor/Node toolchain
all lives in a container, driven by the `./stack` script.

## Contents

- [What's here](#whats-here)
- [The pipeline](#the-pipeline)
- [Prerequisites](#prerequisites)
- [Walkthrough: build, deploy, and talk to the counter](#walkthrough-build-deploy-and-talk-to-the-counter)
- [Understanding the program](#understanding-the-program)
- [Writing your own program](#writing-your-own-program)
- [Command reference](#command-reference)
- [Troubleshooting](#troubleshooting)
- [Stack & versions](#stack--versions)
- [Learn more](#learn-more)

## What's here

| Path | What it is |
|------|------------|
| [`program/`](program/) | An [Anchor](https://www.anchor-lang.com) program (`counter`): a per-wallet PDA you `initialize` then `increment`. Plus fast (LiteSVM) and integration tests. |
| [`clients/ts/`](clients/ts/) | A `@solana/kit` TypeScript client **generated from the program's IDL** by [Codama](https://github.com/codama-idl/codama) — committed under `src/generated/`, regenerable any time. Plus a standalone example. |
| [`app/`](app/) | A Vite + React demo that connects a wallet and drives the program through the generated client. |
| [`docs/`](docs/) | [concepts](docs/concepts.md) · [writing-a-program](docs/writing-a-program.md) · [development](docs/development.md) · [deployment](docs/deployment.md) · [integration](docs/integration.md) |
| `stack` | One entrypoint wrapping Docker. Run `./stack` to list commands. |

## The pipeline

```
program/src/lib.rs  ──anchor build──▶  target/idl/counter.json
        │                                      │
        │                                  codama (./stack codegen)
        ▼                                      ▼
   on-chain program                clients/ts/src/generated/  (typed @solana/kit client)
                                          │              │
                                     app/ (React)   examples/ + tests (standalone TS)
```

Change an instruction's signature in `lib.rs`, rebuild + regen, and the
TypeScript types change with it — call sites stop compiling until you fix them.

## Prerequisites

- **Docker** (with Compose). That's the only hard requirement. The first
  `./stack` command pulls/builds a ~2 GB toolbox image; subsequent runs are fast.
- **A browser wallet** (Phantom, Solflare, Backpack, …) — only for the frontend
  step. Everything else works without one.
- No local Rust, Solana CLI, Anchor, or Node needed.

## Walkthrough: build, deploy, and talk to the counter

Run everything from the `solana/` directory. Each step says what it does and
what you should see.

### 0. One-time setup

```bash
cd solana
cp .env.example .env
```

The defaults target a local validator; nothing here is secret.

### 1. Build the program

```bash
./stack build
```

This compiles `program/programs/counter/src/lib.rs` to a Solana program and
emits its **IDL** at `target/idl/counter.json` — the machine-readable
description of the program's instructions, accounts, and errors. (First run is
slow as it compiles Anchor; later builds are cached.)

### 2. Generate the TypeScript client

```bash
./stack codegen
```

Codama reads the IDL and writes a typed `@solana/kit` client into
[`clients/ts/src/generated/`](clients/ts/src/generated/): instruction builders,
account decoders, and a PDA finder. This is the file set the frontend and tests
import. It's committed for reference but is fully regenerable — never edit it by
hand.

### 3. Run the tests

```bash
./stack test-unit   # fast, in-process (LiteSVM) — milliseconds
./stack test        # integration: deploys to a local validator, drives it via the generated client
```

`./stack test` builds, regenerates the client, spins up a validator, deploys,
and runs the suite in `program/tests/counter.ts` — so it exercises the entire
program → IDL → client → RPC path. You should see all tests pass.

### 4. Start a local validator

```bash
./stack localnet
```

Leave this running in one terminal. It starts **surfpool**, a fast local
validator (RPC on `:8899`, WebSocket on `:8900`, a Studio dashboard on `:18488`).

### 5. Deploy the program to it

In a second terminal:

```bash
./stack deploy-localnet
```

This funds a local wallet and deploys the built program to your running
surfpool. You should see `Deploy success`.

### 6. Drive it with no frontend (optional but instructive)

```bash
./stack example
```

Runs [`clients/ts/examples/increment.ts`](clients/ts/examples/increment.ts): a
plain Node script that creates a wallet, initializes a counter, increments it
twice, and prints `count = 2`. Proof that the generated client works anywhere —
the frontend isn't special.

### 7. Run the frontend and click the button

```bash
./stack app
```

Open **http://localhost:5173**, click **Select Wallet**, and connect a browser
wallet. Then **Create my counter** and **Increment**.

> **Point your wallet at the same cluster the app uses.** With the defaults the
> app talks to your local surfpool (`http://localhost:8899`). In your wallet,
> add/select a custom RPC of `http://localhost:8899`, or — simpler — run the app
> against devnet instead: set `VITE_RPC_URL=https://api.devnet.solana.com` in
> `.env` and use a devnet wallet (after `./stack deploy-devnet`, step 8).

What happens when you click Increment: the app builds an `increment` instruction
with the generated client, the wallet signs and submits it (via the
wallet→kit bridge in [`app/src/solana/useKitSigner.ts`](app/src/solana/useKitSigner.ts)),
and the page re-reads your counter PDA and shows the new value.

### 8. Deploy to a public cluster

```bash
./stack deploy-devnet      # also: deploy-testnet / deploy-mainnet
```

See [docs/deployment.md](docs/deployment.md) for funding, upgrade authority,
priority fees, and verifiable builds before mainnet.

## Understanding the program

Open [`program/programs/counter/src/lib.rs`](program/programs/counter/src/lib.rs).
It's ~90 lines and deliberately minimal. In short:

- A **`Counter` account** holds `{ authority, count, bump }`. It's a separate
  on-chain account (not state "inside" the program).
- It lives at a **PDA** derived from `["counter", authority]`, so each wallet has
  exactly one counter at a predictable address — no lookups.
- Two **instructions**: `initialize` (create the PDA at 0) and `increment` (+1).
- **Account constraints** (`seeds`, `bump`, `has_one = authority`, `Signer`)
  enforce that only a counter's owner can change it — before your code runs.

For a from-scratch explanation of accounts, PDAs, instructions, rent,
discriminators, transactions, and the IDL → client pipeline — all grounded in
this program — read **[docs/concepts.md](docs/concepts.md)**.

## Writing your own program

The change loop is always: **edit `lib.rs` → `./stack build` → `./stack codegen`**,
then use the new generated types. **[docs/writing-a-program.md](docs/writing-a-program.md)**
walks through it with worked, verified examples: adding an instruction (with and
without arguments), adding a field to an account, custom errors and constraints,
testing the change, and scaffolding a brand-new program.

## Command reference

Run `./stack` with no arguments to print this list.

| Command | What it does |
|---------|--------------|
| `./stack build` | Compile the program; emit the IDL. |
| `./stack codegen` | Regenerate the `@solana/kit` client from the IDL. |
| `./stack test` | Build + codegen + integration tests (validator + generated client). |
| `./stack test-unit` | Fast in-process tests (LiteSVM). |
| `./stack localnet` | Run surfpool (local validator) in the foreground. |
| `./stack deploy-localnet` | Airdrop + deploy to the running surfpool. |
| `./stack deploy-devnet` / `-testnet` / `-mainnet` | Deploy to a public cluster. |
| `./stack example` | Run the standalone (no-frontend) TS example. |
| `./stack app` | Run surfpool + the React demo at :5173. |
| `./stack keys-sync` | Sync `declare_id!` / `Anchor.toml` to the program keypair. |
| `./stack verify` | Deterministic verifiable build (installs `solana-verify`). |
| `./stack shell` | Open a shell in the toolbox container. |

## Troubleshooting

- **Wallet shows nothing / transaction fails in the app:** your wallet is on a
  different cluster than the app. Match `VITE_RPC_URL` (see step 7).
- **`anchor test` / surfpool errors about a datasource:** surfpool forks mainnet
  on demand, so the first connection wants internet access. The counter demo
  itself needs nothing off-chain.
- **Generated files are owned by root:** the container runs as root; `./stack
  codegen` chowns the output back to you afterwards.
- **Apple Silicon (arm64):** SBF/verifiable builds are x86-centric and run under
  emulation (slower). If a build misbehaves, set
  `DOCKER_DEFAULT_PLATFORM=linux/amd64`.
- **Re-keyed the program?** Re-run `./stack codegen` (the program id is baked
  into the generated client from the IDL) and update `VITE_PROGRAM_ID` in `.env`.

## Stack & versions

- **Anchor 1.0** on the Agave/Solana 3.x toolchain (Rust program + first-class IDL).
- **Codama** → **`@solana/kit`** for the generated client (the modern,
  tree-shakeable successor to web3.js v1).
- **`@solana/wallet-adapter-react`** for wallet connection. wallet-adapter is
  web3.js-v1-oriented while the client is kit-native, so the app includes a small
  documented bridge ([`app/src/solana/useKitSigner.ts`](app/src/solana/useKitSigner.ts)).
- **surfpool** (LiteSVM-backed) as the local validator — what Anchor 1.0's
  `anchor test` spawns.

> The dev program keypair (`program/counter-keypair.json`) is committed on
> purpose so the repo clones-and-runs. It is a throwaway — generate a fresh one
> for any real deployment (see [docs/deployment.md](docs/deployment.md)).

## Learn more

- **[docs/concepts.md](docs/concepts.md)** — Solana from scratch, via the counter.
- **[docs/writing-a-program.md](docs/writing-a-program.md)** — make changes and add your own program.
- **[docs/development.md](docs/development.md)** — the containers and the dev loop in depth.
- **[docs/deployment.md](docs/deployment.md)** — devnet/testnet/mainnet + verifiable builds.
- **[docs/integration.md](docs/integration.md)** — folding this into the main Go/React app.
