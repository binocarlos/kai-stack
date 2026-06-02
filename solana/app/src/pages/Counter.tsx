import { useCallback, useEffect, useState } from "react";
import {
  appendTransactionMessageInstruction,
  createTransactionMessage,
  pipe,
  setTransactionMessageFeePayerSigner,
  setTransactionMessageLifetimeUsingBlockhash,
  signAndSendTransactionMessageWithSigners,
  type Instruction,
  type TransactionSendingSigner,
} from "@solana/kit";
import {
  fetchMaybeCounter,
  findCounterPda,
  getIncrementInstructionAsync,
  getInitializeInstructionAsync,
} from "@counter-client";
import { rpc } from "../solana/client";
import { useWalletSigner } from "../solana/useKitSigner";

// Build a one-instruction transaction with the wallet as fee payer, then sign
// and send it through the wallet (see useWalletSigner for the kit bridge).
async function sendWithWallet(signer: TransactionSendingSigner, ix: Instruction) {
  const { value: latestBlockhash } = await rpc.getLatestBlockhash().send();
  const message = pipe(
    createTransactionMessage({ version: 0 }),
    (m) => setTransactionMessageFeePayerSigner(signer, m),
    (m) => setTransactionMessageLifetimeUsingBlockhash(latestBlockhash, m),
    (m) => appendTransactionMessageInstruction(ix, m),
  );
  await signAndSendTransactionMessageWithSigners(message);
}

export function Counter() {
  const signer = useWalletSigner();
  const [count, setCount] = useState<bigint | null>(null);
  const [exists, setExists] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!signer) {
      setCount(null);
      setExists(false);
      return;
    }
    const [pda] = await findCounterPda({ authority: signer.address });
    const account = await fetchMaybeCounter(rpc, pda);
    setExists(account.exists);
    setCount(account.exists ? account.data.count : null);
  }, [signer]);

  useEffect(() => {
    refresh().catch((e) => setError(String(e)));
  }, [refresh]);

  const run = useCallback(
    async (build: (s: TransactionSendingSigner) => Promise<Instruction>) => {
      if (!signer) return;
      setBusy(true);
      setError(null);
      try {
        await sendWithWallet(signer, await build(signer));
        // Give the cluster a moment to commit, then re-read the account.
        await new Promise((r) => setTimeout(r, 1500));
        await refresh();
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      } finally {
        setBusy(false);
      }
    },
    [signer, refresh],
  );

  if (!signer) {
    return <p className="hint">Connect a wallet to view and change your counter.</p>;
  }

  return (
    <section className="counter">
      <div className="counter__value">{exists ? String(count) : "—"}</div>

      {exists ? (
        <button
          disabled={busy}
          onClick={() => run((s) => getIncrementInstructionAsync({ authority: s }))}
        >
          {busy ? "Sending…" : "Increment"}
        </button>
      ) : (
        <button
          disabled={busy}
          onClick={() => run((s) => getInitializeInstructionAsync({ authority: s }))}
        >
          {busy ? "Creating…" : "Create my counter"}
        </button>
      )}

      {error && <p className="error">{error}</p>}
    </section>
  );
}
