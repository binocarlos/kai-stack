import { createSolanaRpc } from "@solana/kit";

// The program id and RPC come from Vite env (see .env.example). The browser
// talks to the API/validator over plain HTTP RPC; for sending we rely on the
// wallet (see useKitSigner), so no WebSocket subscriptions are needed here.
export const RPC_URL = import.meta.env.VITE_RPC_URL ?? "http://127.0.0.1:8899";

export const rpc = createSolanaRpc(RPC_URL);
