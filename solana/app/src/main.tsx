import { Buffer } from "buffer";
// @solana/web3.js (v1, pulled in by wallet-adapter) expects a global Buffer.
globalThis.Buffer = globalThis.Buffer ?? Buffer;

import React from "react";
import ReactDOM from "react-dom/client";
import "@solana/wallet-adapter-react-ui/styles.css";
import { App } from "./App";
import { SolanaProviders } from "./solana/WalletProvider";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <SolanaProviders>
      <App />
    </SolanaProviders>
  </React.StrictMode>,
);
