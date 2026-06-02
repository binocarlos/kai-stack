// The wallet-adapter <-> @solana/kit bridge.
//
// wallet-adapter is web3.js-v1-oriented: it hands us a v1 PublicKey and a
// `sendTransaction(v1Tx, connection)` method. The generated Codama client is
// kit-native and builds kit transactions. They don't compose directly, so this
// hook adapts the connected wallet into a kit `TransactionSendingSigner`:
//
//   kit compiled tx --(messageBytes)--> v1 VersionedTransaction
//        --> wallet.sendTransaction (wallet signs + submits) --> signature
//        --> back to kit as SignatureBytes
//
// This is the one rough edge of the wallet-adapter + kit combination. A
// kit-native connector (e.g. ConnectorKit) would remove the v1 hop entirely.
import { useMemo } from "react";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { VersionedMessage, VersionedTransaction } from "@solana/web3.js";
import {
  address,
  getBase58Encoder,
  type SignatureBytes,
  type TransactionSendingSigner,
} from "@solana/kit";

export function useWalletSigner(): TransactionSendingSigner | null {
  const { connection } = useConnection();
  const { publicKey, sendTransaction, connected } = useWallet();

  return useMemo(() => {
    if (!connected || !publicKey || !sendTransaction) return null;

    return {
      address: address(publicKey.toBase58()),
      async signAndSendTransactions(transactions) {
        const signatures: SignatureBytes[] = [];
        for (const tx of transactions) {
          // Rebuild the kit-compiled message as a v1 transaction the wallet
          // understands, then let the wallet sign and submit it.
          const message = VersionedMessage.deserialize(
            new Uint8Array(tx.messageBytes),
          );
          const vtx = new VersionedTransaction(message);
          const signature = await sendTransaction(vtx, connection);
          signatures.push(getBase58Encoder().encode(signature) as SignatureBytes);
        }
        return signatures;
      },
    };
  }, [connected, publicKey, sendTransaction, connection]);
}
