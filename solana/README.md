# Solana reference app

A self-contained, Docker-first blueprint for the full Solana loop: **write a
program → generate typed TypeScript from it → call it from a React frontend (and
any standalone TS) → deploy to devnet/mainnet.**

It lives in its own folder and depends on nothing else in this repo. You need
only **Docker** — the Rust/Solana/Anchor/Node toolchain all lives in a container.

## What's here

| Path | What it is |
|------|------------|
| `program/` | An [Anchor](https://www.anchor-lang.com) program (`counter`): a per-wallet PDA you `initialize` then `increment`. Plus integration tests. |
| `clients/ts/` | A `@solana/kit` TypeScript client **generated from the program's IDL** by [Codama](https://github.com/codama-idl/codama) — committed under `src/generated/`, regenerable any time. Plus a standalone example. |
| `app/` | A Vite + React demo that connects a wallet and drives the program through the generated client. |
| `docs/` | [development](docs/development.md) · [deployment](docs/deployment.md) · [integration](docs/integration.md) |
| `stack` | One entrypoint wrapping Docker (mirrors the repo's root `./stack`). |

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

## Quickstart

```bash
cd solana
cp .env.example .env            # defaults work for local dev

./stack build                   # compile the program (emits the IDL)
./stack codegen                 # generate the @solana/kit client from the IDL
./stack test                    # build + codegen + run integration tests

# Try it for real, locally:
./stack localnet                # terminal 1: a local validator (surfpool)
./stack deploy-localnet         # terminal 2: deploy the program to it
./stack app                     # open http://localhost:5173, connect a wallet
```

`./stack` with no args lists every command. See [docs/development.md](docs/development.md)
for the full local loop and [docs/deployment.md](docs/deployment.md) for devnet/mainnet.

## Stack & versions

- **Anchor 1.0** on the Agave/Solana 3.x toolchain (Rust program + first-class IDL).
- **Codama** → **`@solana/kit`** for the generated client (the modern, tree-shakeable successor to web3.js v1).
- **`@solana/wallet-adapter-react`** for wallet connection. Note: wallet-adapter is web3.js-v1-oriented while the client is kit-native, so the app includes a small documented bridge (`app/src/solana/useKitSigner.ts`).
- **surfpool** (LiteSVM-backed) as the local validator — what Anchor 1.0's `anchor test` spawns.

> The dev program keypair (`program/counter-keypair.json`) is committed on
> purpose so the repo clones-and-runs. It is a throwaway — generate a fresh one
> for any real deployment (see [docs/deployment.md](docs/deployment.md)).
