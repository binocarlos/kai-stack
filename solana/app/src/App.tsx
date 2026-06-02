import { WalletMultiButton } from "@solana/wallet-adapter-react-ui";
import { Counter } from "./pages/Counter";
import { COUNTER_PROGRAM_ADDRESS } from "@counter-client";
import { RPC_URL } from "./solana/client";

export function App() {
  return (
    <main className="app">
      <header className="app__header">
        <h1>Solana Counter</h1>
        <WalletMultiButton />
      </header>

      <Counter />

      <footer className="app__footer">
        <div>program: <code>{COUNTER_PROGRAM_ADDRESS}</code></div>
        <div>rpc: <code>{RPC_URL}</code></div>
      </footer>
    </main>
  );
}
