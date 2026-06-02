//! Counter — the minimal reference Solana program.
//!
//! It demonstrates the load-bearing concepts a real program uses, with nothing
//! else in the way:
//!   - a Program Derived Address (PDA) per authority, so each wallet owns its
//!     own counter at a deterministic address (no account to pass around);
//!   - `init` with rent-exempt sizing via `#[derive(InitSpace)]`;
//!   - `has_one` authority checks so only the owner can increment;
//!   - a typed custom error.
//!
//! Everything downstream (the IDL, the generated `@solana/kit` client, the
//! React app) is derived from the interface defined here. Change an instruction
//! signature and the regenerated TypeScript types change with it.

use anchor_lang::prelude::*;

// Program id. Matches the committed dev keypair (program/counter-keypair.json),
// which is mounted into the build so `git clone && ./stack build` just works.
// For a REAL devnet/mainnet deploy, generate a fresh program keypair and run
// `./stack keys-sync` — see docs/deployment.md. Never reuse this dev identity.
declare_id!("8yy1KCyWhqUxgebL6yvZeiA7NwLRQnEuxj9AoSSvYKTR");

#[program]
pub mod counter {
    use super::*;

    /// Create the caller's counter PDA, owned by `authority`, starting at 0.
    pub fn initialize(ctx: Context<Initialize>) -> Result<()> {
        let counter = &mut ctx.accounts.counter;
        counter.authority = ctx.accounts.authority.key();
        counter.count = 0;
        counter.bump = ctx.bumps.counter;
        msg!("counter initialized for {}", counter.authority);
        Ok(())
    }

    /// Add one to the caller's counter. Only the authority may do this.
    pub fn increment(ctx: Context<Increment>) -> Result<()> {
        let counter = &mut ctx.accounts.counter;
        counter.count = counter
            .count
            .checked_add(1)
            .ok_or(CounterError::Overflow)?;
        msg!("counter is now {}", counter.count);
        Ok(())
    }
}

#[derive(Accounts)]
pub struct Initialize<'info> {
    #[account(
        init,
        payer = authority,
        space = 8 + Counter::INIT_SPACE,
        seeds = [b"counter", authority.key().as_ref()],
        bump
    )]
    pub counter: Account<'info, Counter>,
    #[account(mut)]
    pub authority: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Increment<'info> {
    #[account(
        mut,
        seeds = [b"counter", authority.key().as_ref()],
        bump = counter.bump,
        has_one = authority
    )]
    pub counter: Account<'info, Counter>,
    pub authority: Signer<'info>,
}

/// One counter, owned by `authority`. `INIT_SPACE` (from `#[derive(InitSpace)]`)
/// sums the field sizes; the +8 in `space` above is the Anchor account
/// discriminator.
#[account]
#[derive(InitSpace)]
pub struct Counter {
    pub authority: Pubkey,
    pub count: u64,
    pub bump: u8,
}

#[error_code]
pub enum CounterError {
    #[msg("counter overflowed u64")]
    Overflow,
}
