# Development

Everything runs in Docker. The only prerequisite is Docker (with Compose). The
`./stack` script wraps it all — run `./stack` with no arguments to list commands.

## The containers

`docker-compose.yml` defines three services:

- **`toolbox`** — the official `solanafoundation/anchor:v1.0.2` image + Node 22 +
  the `surfpool` binary (see `docker/toolbox.Dockerfile`). Every `anchor`,
  `solana`, and codegen command runs here, so you never install a local
  toolchain. Built once on first use; rebuilt with `docker compose build toolbox`.
- **`surfpool`** — the local validator (LiteSVM-backed, near-instant boot). Used
  for interactive dev and by the app. `anchor test` spawns its own surfpool.
- **`app`** — the Vite dev server (`node:22-alpine`).

Docker volumes persist the cargo registry, the program `target/`, and your
`~/.config/solana` keypairs across runs so rebuilds stay fast.

## The inner loop

```bash
./stack build        # anchor build -> target/idl/counter.json + the .so
./stack codegen      # IDL -> clients/ts/src/generated (the @solana/kit client)
./stack test         # build + codegen + integration tests (anchor test + surfpool)
```

`./stack test` runs the suite in `program/tests/counter.ts` against a validator,
using the generated client — so it exercises the whole program→IDL→types→RPC
path, not just the program in isolation.

### Editing the program

`program/programs/counter/src/lib.rs` is the source of truth for the on-chain
interface. After changing an instruction:

```bash
./stack build && ./stack codegen
```

The regenerated client in `clients/ts/src/generated/` reflects the new
signature; any TypeScript that called the old shape stops compiling. The
generated code is committed (as reference) but is fully disposable — never edit
it by hand.

## Running it locally end to end

```bash
./stack localnet            # terminal 1: start surfpool (RPC :8899, WS :8900, Studio :18488)
./stack deploy-localnet     # terminal 2: airdrop + deploy the program to surfpool
./stack example             # standalone Node script: init + increment, no frontend
./stack app                 # the React app at http://localhost:5173
```

`./stack example` runs `clients/ts/examples/increment.ts` — proof that the
generated client works in plain Node, identically to how the frontend uses it.

## The frontend

`./stack app` brings up surfpool + the Vite dev server. Open
http://localhost:5173, connect a browser wallet (Phantom/Solflare/…; they're
auto-discovered via the Solana Wallet Standard), and Create / Increment your
counter.

> Point the wallet at the same cluster the app uses (`VITE_RPC_URL`). For
> localnet that means adding a custom RPC (`http://localhost:8899`) in the
> wallet, or just run the app against devnet by setting `VITE_RPC_URL` in `.env`.

The wallet→kit bridge lives in `app/src/solana/useKitSigner.ts`: wallet-adapter
is web3.js-v1-oriented and the generated client is kit-native, so the hook
adapts the connected wallet into a kit `TransactionSendingSigner`. It's the one
awkward seam of this library combination and is commented in full.

## Notes & gotchas

- **arm64 / Apple Silicon:** the SBF and verifiable-build toolchains are
  x86-centric. On Apple Silicon, builds run under emulation (slower); if a build
  misbehaves, force the platform: `DOCKER_DEFAULT_PLATFORM=linux/amd64`.
- **Generated files are root-owned by default** (the container runs as root);
  `./stack codegen` chowns them back to you afterwards.
- **`surfpool` forks mainnet on demand**, so the first connection wants network
  access. The counter demo itself needs nothing off-chain.
- **Bundle size:** the demo app isn't tree-shaking-tuned; that's fine for a
  reference. web3.js v1 (via wallet-adapter) is the bulk of it.
