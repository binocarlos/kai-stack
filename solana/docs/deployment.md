# Deployment

Deploying to a public cluster differs from localnet in two ways that matter:
**you need a real, funded keypair**, and for mainnet **you want a fresh program
id and a verifiable build**. Everything still runs through the `toolbox`
container.

## Keypairs

There are two distinct keypairs:

| Keypair | What it is | Where |
|---------|------------|-------|
| **Program keypair** | Defines the program id (`declare_id!`). | `program/counter-keypair.json` (committed *dev* key) |
| **Deployer / upgrade authority** | Pays for the deploy and is allowed to upgrade. | `~/.config/solana/id.json` in the toolbox (persisted in a volume) |

> ⚠️ The committed `program/counter-keypair.json` is a **throwaway dev identity**
> so the repo clones-and-runs. For any real deployment, generate a fresh one:
>
> ```bash
> ./stack shell
> solana-keygen new -o program/target/deploy/counter-keypair.json   # new program id
> anchor keys sync                                                   # rewrite declare_id! + Anchor.toml
> exit
> ./stack build && ./stack codegen                                   # rebuild with the new id
> ```
>
> Then remove the committed dev keypair and the gitignore exception for it.

Fund the deployer (devnet/testnet have faucets; mainnet you transfer real SOL):

```bash
./stack shell
solana config set --url devnet
solana airdrop 5            # devnet/testnet only
solana address             # the deployer pubkey
```

## Deploy

```bash
./stack deploy-devnet       # anchor deploy --provider.cluster devnet
./stack deploy-testnet
./stack deploy-mainnet
```

These wrap `anchor deploy`, which uploads to a **buffer account** then sets the
program. The deployer keypair is the upgrade authority by default.

- **Priority fees:** congested clusters (mainnet especially) often need a
  priority fee or the deploy stalls. Add `--with-compute-unit-price <microlamports>`
  via `anchor deploy -- --with-compute-unit-price 50000`, and retry.
- **Stuck deploys leak buffers:** if a deploy fails partway, reclaim the rent:
  `./stack shell` then `solana program close --buffers`.
- **Upgrades:** redeploy with the same program id and upgrade-authority keypair.
  To make a program immutable, set the upgrade authority to `none`.

After deploying, update the app's `.env`:

```
VITE_RPC_URL=https://api.devnet.solana.com
VITE_PROGRAM_ID=<your program id>   # only if you re-keyed
```

(`VITE_PROGRAM_ID` is also baked into the generated client from the IDL, so a
re-key means re-running `./stack codegen`.)

## Verifiable builds (recommended before mainnet)

A verifiable build is deterministic, so anyone can confirm the on-chain bytes
match this source. `solana-verify` isn't in the base image, so `./stack verify`
installs it on first use (cached in the cargo volume):

```bash
./stack verify              # solana-verify build (deterministic, in Docker)
```

Then deploy **that exact artifact** and submit verification:

```bash
./stack shell
solana program deploy target/verifiable/counter.so \
  --program-id program/counter-keypair.json --url mainnet \
  --with-compute-unit-price 50000
solana-verify verify-from-repo <repo-url> --program-id <program id>
```

Do **not** verify an artifact built by plain `anchor build` — only the
`solana-verify build` output hashes match. See the
[Anchor verifiable builds docs](https://www.anchor-lang.com/docs/references/verifiable-builds).

## Cluster cheat sheet

| Cluster | RPC | Faucet |
|---------|-----|--------|
| localnet (surfpool) | `http://localhost:8899` | `solana airdrop` |
| devnet | `https://api.devnet.solana.com` | `solana airdrop` |
| testnet | `https://api.testnet.solana.com` | `solana airdrop` |
| mainnet-beta | a paid RPC (Helius/QuickNode/…) | real SOL |
