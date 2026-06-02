// Small hand-written helpers that sit next to the generated client, for flows
// that sign with a local keypair (the standalone example and the integration
// tests). The frontend signs with a wallet instead and so builds its own send
// path (see app/src/solana/). Everything here is plain @solana/kit.
import {
  airdropFactory,
  appendTransactionMessageInstruction,
  assertIsTransactionWithBlockhashLifetime,
  createSolanaRpc,
  createSolanaRpcSubscriptions,
  createTransactionMessage,
  getSignatureFromTransaction,
  pipe,
  sendAndConfirmTransactionFactory,
  setTransactionMessageFeePayerSigner,
  setTransactionMessageLifetimeUsingBlockhash,
  signTransactionMessageWithSigners,
  type Instruction,
  type TransactionSigner,
} from "@solana/kit";

export type Client = ReturnType<typeof createClient>;

export function createClient(
  rpcUrl = "http://127.0.0.1:8899",
  wsUrl = "ws://127.0.0.1:8900",
) {
  const rpc = createSolanaRpc(rpcUrl);
  const rpcSubscriptions = createSolanaRpcSubscriptions(wsUrl);
  return {
    rpc,
    rpcSubscriptions,
    sendAndConfirm: sendAndConfirmTransactionFactory({ rpc, rpcSubscriptions }),
    airdrop: airdropFactory({ rpc, rpcSubscriptions }),
  };
}

/** Build a single-instruction tx, sign it with `payer`, send and confirm. */
export async function sendInstruction(
  client: Client,
  payer: TransactionSigner,
  ix: Instruction,
): Promise<string> {
  const { value: latestBlockhash } = await client.rpc.getLatestBlockhash().send();
  const message = pipe(
    createTransactionMessage({ version: 0 }),
    (m) => setTransactionMessageFeePayerSigner(payer, m),
    (m) => setTransactionMessageLifetimeUsingBlockhash(latestBlockhash, m),
    (m) => appendTransactionMessageInstruction(ix, m),
  );
  const signed = await signTransactionMessageWithSigners(message);
  // Signing widens the lifetime to a union; narrow it back to blockhash
  // lifetime (which carries lastValidBlockHeight) for the confirmer.
  assertIsTransactionWithBlockhashLifetime(signed);
  await client.sendAndConfirm(signed, { commitment: "confirmed" });
  return getSignatureFromTransaction(signed);
}
