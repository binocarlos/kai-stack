// Standalone proof that the generated client is frontend-independent: a plain
// Node script (no React, no wallet) that drives the program with the SAME
// generated @solana/kit client the frontend imports.
//
//   ./stack localnet            # terminal 1: start surfpool
//   ./stack deploy-localnet      # once, so the program is deployed locally
//   yarn example:increment      # terminal 2, from clients/ts
//
// RPC_URL/WS_URL default to a local validator; point them at devnet to run there.
import { generateKeyPairSigner, lamports } from "@solana/kit";
import { createClient, sendInstruction } from "../src/rpc";
import {
  fetchCounter,
  findCounterPda,
  getIncrementInstructionAsync,
  getInitializeInstructionAsync,
} from "../src/generated";

const client = createClient(
  process.env.RPC_URL ?? "http://127.0.0.1:8899",
  process.env.WS_URL ?? "ws://127.0.0.1:8900",
);

// A throwaway authority so the demo always starts from a fresh counter.
const authority = await generateKeyPairSigner();
console.log("authority:", authority.address);

await client.airdrop({
  recipientAddress: authority.address,
  lamports: lamports(1_000_000_000n),
  commitment: "confirmed",
});

// `...Async` builders resolve the counter PDA from the authority for us.
await sendInstruction(client, authority, await getInitializeInstructionAsync({ authority }));
await sendInstruction(client, authority, await getIncrementInstructionAsync({ authority }));
await sendInstruction(client, authority, await getIncrementInstructionAsync({ authority }));

const [counterPda] = await findCounterPda({ authority: authority.address });
const counter = await fetchCounter(client.rpc, counterPda);
console.log(`counter ${counterPda} -> count = ${counter.data.count}`);
