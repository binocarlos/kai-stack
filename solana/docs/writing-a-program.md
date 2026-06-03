# Writing a program

This walks you through changing the program and watching the change flow all the
way to the frontend. If the concepts feel unfamiliar, read
[concepts.md](concepts.md) first. Every example below has been run against this
repo, so you can follow along verbatim.

## The change loop

Whenever you touch the program, the loop is always the same three steps:

```bash
# 1. edit program/programs/counter/src/lib.rs
./stack build      # 2. recompile -> new target/idl/counter.json
./stack codegen    # 3. regenerate clients/ts/src/generated from the new IDL
```

After step 3, the new TypeScript types exist and any caller (tests, the example,
the app) can use them. Old call sites that no longer match stop compiling — that
is the safety net working.

> The program id doesn't change when you edit instructions, so you don't need to
> redeploy keys or touch `Anchor.toml`. You only re-key when starting a brand-new
> program (see the bottom of this page) or deploying for real
> ([deployment.md](deployment.md)).

## Example 1 — add an instruction (no arguments): `reset`

Add a `reset` that sets the counter back to zero. It needs the same accounts and
checks as `increment` (mutable counter PDA, signed by its authority), so we
reuse the existing `Increment` accounts context.

In `program/programs/counter/src/lib.rs`, inside `mod counter`, add:

```rust
/// Set the caller's counter back to zero. Only the authority may do this.
pub fn reset(ctx: Context<Increment>) -> Result<()> {
    ctx.accounts.counter.count = 0;
    Ok(())
}
```

Then:

```bash
./stack build && ./stack codegen
```

A new file `clients/ts/src/generated/instructions/reset.ts` appears, exporting
`getResetInstructionAsync` and `getResetInstruction` — same shape as the
increment builders. Use it exactly like increment:

```ts
// in a test or script
await sendInstruction(client, authority, await getResetInstructionAsync({ authority }));
```

```tsx
// in the app: another button on the Counter page
<button onClick={() => run((s) => getResetInstructionAsync({ authority: s }))}>
  Reset
</button>
```

That's the whole loop for a new instruction: write the Rust handler, rebuild,
regen, call the generated builder.

## Example 2 — an instruction with arguments: `incrementBy`

Instructions can take typed arguments. Add one that increments by an amount:

```rust
pub fn increment_by(ctx: Context<Increment>, amount: u64) -> Result<()> {
    let counter = &mut ctx.accounts.counter;
    counter.count = counter
        .count
        .checked_add(amount)
        .ok_or(CounterError::Overflow)?;
    Ok(())
}
```

After `./stack build && ./stack codegen`, the generated builder gains the
argument as a typed field — the Rust `amount: u64` becomes a `bigint` you pass in:

```ts
await getIncrementByInstructionAsync({ authority, amount: 5n });
```

So: **accounts** in the `#[derive(Accounts)]` struct become accounts in the
builder input; **function arguments** become typed fields in the same input.
Anchor serialises them (Borsh); Codama generates the matching encoder/decoder.

## Example 3 — add a field to an account

Suppose each counter should remember when it was last touched. Add a field to
the `Counter` struct:

```rust
#[account]
#[derive(InitSpace)]
pub struct Counter {
    pub authority: Pubkey,
    pub count: u64,
    pub bump: u8,
    pub last_slot: u64,     // new
}
```

Set it in the handlers (e.g. `counter.last_slot = Clock::get()?.slot;`). Two
things to know:

- `#[derive(InitSpace)]` recomputes the account size automatically, so
  `space = 8 + Counter::INIT_SPACE` stays correct for newly created accounts.
- **Existing accounts don't grow.** Adding a field changes the layout, so
  counters created before the change won't decode against the new type. On
  localnet just wipe state (restart `./stack localnet`); on devnet/mainnet you'd
  need a migration/realloc strategy. For learning, develop on a fresh localnet.

After rebuild + codegen, `fetchCounter(...)` returns the new field, typed.

## Writing useful checks and errors

- **Custom errors** live in the `#[error_code]` enum. Add a variant, return it
  with `.ok_or(CounterError::Whatever)?` or `require!(cond, CounterError::Whatever)`.
  It shows up in the IDL and as a typed constant in `clients/ts/src/generated/errors/`.
- **Account constraints** (in the `#[account(...)]` attributes) are your first
  line of defence: `mut`, `has_one`, `seeds`/`bump`, `init`, `constraint = ...`.
  Prefer a constraint over a manual `if` — it runs before your code and reads
  clearly in the IDL.
- **Always test the unhappy path.** The counter's third test asserts a wrong
  signer is rejected; copy that pattern for your new rules.

## Testing your change

Two layers, both already wired up (see [development.md](development.md)):

- **Fast unit tests** — `program/tests-litesvm/counter.litesvm.ts`. Run with
  `./stack test-unit`. In-process, milliseconds, great for iterating on logic.
- **Integration tests** — `program/tests/counter.ts`. Run with `./stack test`.
  Deploys to a real local validator and exercises the generated client over RPC.

Add a test for your new instruction in whichever layer fits, mirroring the
existing cases.

## Starting a brand-new program (not editing the counter)

When you want a second, independent program rather than extending the counter:

1. **Scaffold** a new crate under `program/programs/<name>/` with its own
   `Cargo.toml` (copy the counter's) and `src/lib.rs`.
2. **Give it an id:** `./stack shell`, then `anchor keys list` / generate a
   keypair, and add the program under `[programs.localnet]` in `Anchor.toml`.
   Run `anchor keys sync` so `declare_id!` matches.
3. **Build:** `./stack build` compiles all programs in the workspace and emits an
   IDL per program (`target/idl/<name>.json`).
4. **Generate a client:** the codegen step
   ([`clients/ts/generate.mjs`](../clients/ts/generate.mjs)) currently points at
   `counter.json`. Add a second `createFromRoot(...).accept(renderVisitor(...))`
   block for your new IDL, writing to its own output folder (e.g.
   `clients/ts/src/generated-<name>/`), then `./stack codegen`.
5. **Deploy** it alongside the counter with the same `./stack deploy-*` commands
   (they deploy the whole workspace).

For real deployments and verifiable builds, see [deployment.md](deployment.md).
