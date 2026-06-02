# Integrating with the main stack

This folder is deliberately standalone: it has its own `docker-compose.yml`,
`stack` script, and toolchain, and touches nothing else in the repo. That keeps
the Solana concern isolated while it stabilises. When you're ready to fold it
into the main app (Go API + `frontend/`), here's the shape of the work. None of
it is done yet — this is a map, not a checklist of completed steps.

## What plugs in where

**1. The generated client → the main frontend.**
`clients/ts/src/generated` is a plain `@solana/kit` client with no app-specific
assumptions. The repo's `frontend/` can consume it the same way `app/` does:
add `@solana/kit` (+ `@solana/program-client-core`) and a wallet connector, then
import the generated instruction/account/PDA helpers. The wallet→kit bridge in
`app/src/solana/useKitSigner.ts` ports over directly.

- Decide whether the browser signs (wallet-adapter, as here) or whether the Go
  API builds/sponsors transactions and the browser only signs the result.
- `frontend/` uses MUI + TanStack Query; wrap the counter reads in a `useQuery`
  hook and the writes in `useMutation`, mirroring `frontend/src/hooks/`.

**2. Codegen → the repo's type-sync convention.**
The root `./stack generate_types` regenerates Go→TS types. Add a sibling step
that runs `./stack codegen` (or call it from CI) so the on-chain client stays in
lockstep with the program, the same way `frontend/src/types/gotypes.ts` tracks
the Go models.

**3. The program lifecycle → the Go API (optional).**
If the backend needs to read on-chain state or submit transactions, generate a
**Rust** client from the same IDL (`@codama/renderers-rust`) for a Go↔Rust
sidecar, or call the program over JSON-RPC from Go directly. The IDL
(`program/target/idl/counter.json`) is the shared contract either way.

**4. Compose → the root stack.**
The root `docker-compose.yml` fronts services with the `noxy` router. To run the
demo app under the same router, port the `app` service definition over and add a
route; the `toolbox`/`surfpool` services are dev-only and can stay here.

## What to keep separate

- The **Rust/Anchor build toolchain** is heavy (~GBs). Keep it in the `toolbox`
  image rather than the Go/Node images so the main build stays lean.
- **surfpool** is a dev validator — it has no place in a production compose file.
- The committed **dev program keypair** must not travel to production; real
  deployments use their own (see [deployment.md](deployment.md)).

## Suggested first step

Move just the generated client: have `frontend/` import from
`solana/clients/ts/src/generated` behind one hook (e.g. `useCounter`), wire a
wallet, and prove the round trip from the real app against devnet. Everything
else (Go integration, shared compose) can follow once that path is solid.
