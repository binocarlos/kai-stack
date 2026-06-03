# Solana concepts, by way of the counter

If you've never written a Solana program, read this once. It explains the ideas
you need using the actual counter in this repo, then points at the lines that
implement them. Everything here maps to
[`program/programs/counter/src/lib.rs`](../program/programs/counter/src/lib.rs).

## The one big idea: programs and accounts are separate

On most platforms a "smart contract" bundles code *and* its stored data. Solana
splits them:

- **Programs** are stateless code. They are deployed once and don't hold any
  per-user data themselves.
- **Accounts** hold data (and SOL). Every account has an owner program; only the
  owner may change its data.

So "the counter for wallet X" is **not** stored inside the program. It's a
separate **account** that the counter program owns. The program is just the
logic that creates and mutates those accounts.

This is why almost every instruction takes a list of accounts to operate on —
the program needs to be handed the data it's allowed to touch.

## Accounts

An account is a blob of bytes plus some metadata (owner, lamports balance). Our
data account is the `Counter` struct:

```rust
#[account]
#[derive(InitSpace)]
pub struct Counter {
    pub authority: Pubkey,  // who owns this counter
    pub count: u64,         // the number
    pub bump: u8,           // PDA bump (explained below)
}
```

On chain this is stored as: an 8-byte **discriminator** (Anchor writes this so
it can tell account types apart) followed by the fields. That's why the account
size is `8 + Counter::INIT_SPACE` where you see `space = ...` in the code.

**Rent:** accounts must hold enough SOL to be "rent-exempt" or the network
reclaims them. The payer covers this when the account is created (`init`).

## Program Derived Addresses (PDAs)

How do we find *your* counter account without storing its address somewhere? We
derive it deterministically from seeds:

```rust
seeds = [b"counter", authority.key().as_ref()]
```

A PDA is an address computed from `(seeds, program_id)`. It has **no private
key** — instead the program is allowed to "sign" for it. Two consequences:

1. Given a wallet, anyone can recompute its counter address (the client does
   this with `findCounterPda({ authority })`). No lookup table needed.
2. Each wallet gets exactly one counter, at a predictable address.

The `bump` is a small number that nudges the derivation onto a valid address; we
store it so later instructions don't have to recompute it.

## Instructions

Instructions are the program's entry points — its public API. The counter has
two:

```rust
pub fn initialize(ctx: Context<Initialize>) -> Result<()> { ... }
pub fn increment(ctx: Context<Increment>) -> Result<()> { ... }
```

Each takes a `Context<T>` describing the accounts it needs, and optionally some
arguments. The `#[derive(Accounts)]` structs (`Initialize`, `Increment`) declare
those accounts and the rules they must satisfy — Anchor checks the rules before
your function body runs:

```rust
#[derive(Accounts)]
pub struct Increment<'info> {
    #[account(
        mut,                                              // will be modified
        seeds = [b"counter", authority.key().as_ref()],   // must be THIS pda
        bump = counter.bump,
        has_one = authority                               // counter.authority == authority
    )]
    pub counter: Account<'info, Counter>,
    pub authority: Signer<'info>,                          // must sign the tx
}
```

`Signer` means that account must have signed the transaction. `has_one =
authority` means the stored `counter.authority` must equal the passed
`authority`. Together they enforce "only the owner can increment their counter"
— see the third test in
[`tests-litesvm/counter.litesvm.ts`](../program/tests-litesvm/counter.litesvm.ts),
which proves a different signer is rejected.

## Discriminators

Each instruction and account type has an 8-byte tag (the first bytes of a hash
of its name). The program uses it to dispatch the right instruction; clients use
it to recognise account types. You never write these by hand — Anchor generates
them and they appear in the IDL.

## Transactions

A client doesn't call an instruction directly; it builds a **transaction**: one
or more instructions, a fee payer, a recent blockhash (so the tx expires), and
signatures. The validator runs the instructions atomically — all succeed or none
do. In TypeScript with `@solana/kit` that looks like:

```ts
const message = pipe(
  createTransactionMessage({ version: 0 }),
  (m) => setTransactionMessageFeePayerSigner(authority, m),
  (m) => setTransactionMessageLifetimeUsingBlockhash(latestBlockhash, m),
  (m) => appendTransactionMessageInstruction(incrementIx, m),
);
const signed = await signTransactionMessageWithSigners(message);
```

(See [`clients/ts/src/rpc.ts`](../clients/ts/src/rpc.ts) for the full helper.)

## The IDL: the contract between program and clients

When you run `anchor build`, Anchor emits **`target/idl/counter.json`** — a
machine-readable description of every instruction, account, error, and PDA.
This is the single source of truth that ties the on-chain program to the
off-chain code.

## The generated client

You don't write the TypeScript bindings by hand. [Codama](https://github.com/codama-idl/codama)
reads the IDL and generates a typed `@solana/kit` client into
[`clients/ts/src/generated/`](../clients/ts/src/generated/):

- `instructions/` — builders like `getIncrementInstructionAsync({ authority })`
  that also resolve the PDA for you.
- `accounts/` — `fetchCounter(rpc, pda)` and decoders that turn raw bytes back
  into a typed `{ authority, count, bump }`.
- `pdas/` — `findCounterPda({ authority })`.
- `programs/` — `COUNTER_PROGRAM_ADDRESS` and helpers.
- `errors/` — your `#[error_code]` variants as typed constants.

Because the client is generated *from* the program, the two can't drift: change
an instruction, rebuild, regenerate, and the TypeScript types change with it.

## Putting it together

```
lib.rs (Rust)  ──anchor build──▶  IDL  ──codama──▶  clients/ts/src/generated  ──▶  app + tests + scripts
   the program                  the API           typed @solana/kit client       things that call it
```

Now read [writing-a-program.md](writing-a-program.md) to make a change and watch
it ripple through this chain.
