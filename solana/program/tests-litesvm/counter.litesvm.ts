// Fast unit tests: load the compiled program into LiteSVM in-process (no
// validator, milliseconds per test). Like tests/counter.ts it drives the
// program through the generated @solana/kit client, but executes against an
// in-memory SVM instead of real RPC — so it's the quick inner-loop check.
//
// Run with: ./stack test-unit   (builds the .so first, then runs this in Node).
import assert from "node:assert/strict";
import { dirname, resolve } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

import { FailedTransactionMetadata, LiteSVM } from "litesvm";
import {
  appendTransactionMessageInstruction,
  blockhash,
  createTransactionMessage,
  generateKeyPairSigner,
  lamports,
  pipe,
  setTransactionMessageFeePayerSigner,
  setTransactionMessageLifetimeUsingBlockhash,
  signTransactionMessageWithSigners,
  type Address,
  type Instruction,
  type KeyPairSigner,
} from "@solana/kit";
import {
  COUNTER_PROGRAM_ADDRESS,
  decodeCounter,
  findCounterPda,
  getIncrementInstruction,
  getIncrementInstructionAsync,
  getInitializeInstructionAsync,
} from "../../clients/ts/src/generated";

const SO_PATH = resolve(
  dirname(fileURLToPath(import.meta.url)),
  "../target/deploy/counter.so",
);

function newSvm(): LiteSVM {
  const svm = new LiteSVM();
  svm.addProgramFromFile(COUNTER_PROGRAM_ADDRESS, SO_PATH);
  return svm;
}

async function fundedSigner(svm: LiteSVM): Promise<KeyPairSigner> {
  const signer = await generateKeyPairSigner();
  svm.airdrop(signer.address, lamports(1_000_000_000n));
  return signer;
}

async function send(
  svm: LiteSVM,
  authority: KeyPairSigner,
  ix: Instruction,
  { expectOk = true } = {},
) {
  // Advance the blockhash so repeated identical instructions (e.g. two
  // increments) produce distinct signatures rather than a duplicate tx.
  svm.expireBlockhash();
  const message = pipe(
    createTransactionMessage({ version: 0 }),
    (m) => setTransactionMessageFeePayerSigner(authority, m),
    (m) =>
      setTransactionMessageLifetimeUsingBlockhash(
        { blockhash: blockhash(svm.latestBlockhash()), lastValidBlockHeight: 2n ** 63n },
        m,
      ),
    (m) => appendTransactionMessageInstruction(ix, m),
  );
  const signed = await signTransactionMessageWithSigners(message);
  const result = svm.sendTransaction(signed);
  const failed = result instanceof FailedTransactionMetadata;
  if (expectOk) assert.ok(!failed, `tx unexpectedly failed: ${failed ? result.err() : ""}`);
  return result;
}

async function readCount(svm: LiteSVM, pda: Address): Promise<bigint> {
  const account = svm.getAccount(pda);
  assert.ok(account, "counter account should exist");
  return decodeCounter(account).data.count;
}

test("initialize then increment updates the count", async () => {
  const svm = newSvm();
  const authority = await fundedSigner(svm);
  const [pda] = await findCounterPda({ authority: authority.address });

  await send(svm, authority, await getInitializeInstructionAsync({ authority }));
  assert.equal(await readCount(svm, pda), 0n);

  await send(svm, authority, await getIncrementInstructionAsync({ authority }));
  await send(svm, authority, await getIncrementInstructionAsync({ authority }));
  assert.equal(await readCount(svm, pda), 2n);
});

test("a wrong authority cannot increment someone else's counter", async () => {
  const svm = newSvm();
  const owner = await fundedSigner(svm);
  await send(svm, owner, await getInitializeInstructionAsync({ authority: owner }));
  const [ownerPda] = await findCounterPda({ authority: owner.address });

  // Attacker signs but targets the owner's PDA: the seeds/has_one constraints
  // (derived from the signer) reject it.
  const attacker = await fundedSigner(svm);
  const badIx = getIncrementInstruction({ counter: ownerPda, authority: attacker });
  const result = await send(svm, attacker, badIx, { expectOk: false });
  assert.ok(result instanceof FailedTransactionMetadata);
});
