import type { FaucetClient, LocalECDSAKeySigner, PublicClient } from "@nilfoundation/niljs";
import { createEvent, createStore } from "../store.ts";

// Public Client
export const setPublicClient = createEvent<PublicClient>();
export const $publicClient = createStore<PublicClient | null>(null).on(
  setPublicClient,
  (_, client) => client,
);

$publicClient.watch((client) => {
  console.log("PublicClient updated:", client);
});

// Signer
export const setSigner = createEvent<LocalECDSAKeySigner>();
export const $signer = createStore<LocalECDSAKeySigner | null>(null).on(
  setSigner,
  (_, signer) => signer,
);

$signer.watch((signer) => {
  console.log("Signer updated:", signer);
});

// Faucet Client
export const setFaucetClient = createEvent<FaucetClient>();
export const $faucetClient = createStore<FaucetClient | null>(null).on(
  setFaucetClient,
  (_, faucetClient) => faucetClient,
);

$faucetClient.watch((faucetClient) => {
  console.log("FaucetClient updated:", faucetClient);
});
