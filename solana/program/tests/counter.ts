// Integration tests for the counter program.
//
// These run under `anchor test` (./stack test): Anchor builds, starts a local
// validator, deploys the program, then runs this file. The suite talks to the
// validator through the GENERATED @solana/kit client — the same client the
// frontend uses — so it dogfoods the whole program -> IDL -> types pipeline.
import assert from "node:assert/strict";
import { test } from "node:test";

import { generateKeyPairSigner, lamports, type KeyPairSigner } from "@solana/kit";

import { createClient, sendInstruction } from "../../clients/ts/src/rpc";
import {
  fetchCounter,
  findCounterPda,
  getIncrementInstruction,
  getIncrementInstructionAsync,
  getInitializeInstructionAsync,
} from "../../clients/ts/src/generated";

const client = createClient(
  process.env.RPC_URL ?? "http://127.0.0.1:8899",
  process.env.WS_URL ?? "ws://127.0.0.1:8900",
);

async function fundedSigner(): Promise<KeyPairSigner> {
  const signer = await generateKeyPairSigner();
  await client.airdrop({
    recipientAddress: signer.address,
    lamports: lamports(1_000_000_000n),
    commitment: "confirmed",
  });
  return signer;
}

test("initialize starts the counter at zero", async () => {
  const authority = await fundedSigner();
  await sendInstruction(client, authority, await getInitializeInstructionAsync({ authority }));

  const [pda] = await findCounterPda({ authority: authority.address });
  const counter = await fetchCounter(client.rpc, pda);
  assert.equal(counter.data.count, 0n);
  assert.equal(counter.data.authority, authority.address);
});

test("increment adds one each time", async () => {
  const authority = await fundedSigner();
  await sendInstruction(client, authority, await getInitializeInstructionAsync({ authority }));
  await sendInstruction(client, authority, await getIncrementInstructionAsync({ authority }));
  await sendInstruction(client, authority, await getIncrementInstructionAsync({ authority }));

  const [pda] = await findCounterPda({ authority: authority.address });
  const counter = await fetchCounter(client.rpc, pda);
  assert.equal(counter.data.count, 2n);
});

test("a different signer cannot touch someone else's counter", async () => {
  const owner = await fundedSigner();
  await sendInstruction(client, owner, await getInitializeInstructionAsync({ authority: owner }));
  const [ownerPda] = await findCounterPda({ authority: owner.address });

  // Attacker signs but points at the owner's PDA. The seeds + has_one
  // constraints are derived from the signer, so the program rejects it.
  const attacker = await fundedSigner();
  const badIx = getIncrementInstruction({ counter: ownerPda, authority: attacker });
  await assert.rejects(() => sendInstruction(client, attacker, badIx));
});
