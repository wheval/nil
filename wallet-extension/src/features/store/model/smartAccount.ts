import type { Hex, SmartAccountV1 } from "@nilfoundation/niljs";
import { combine, sample } from "effector";
import { saveUserDetails } from "../../../background/storage";
import {
  createClient,
  createFaucetClient,
  createSigner,
  initializeOrDeploySmartAccount,
  topUpAllTokens,
} from "../../blockchain";
import { generateRandomShard } from "../../utils";
import { createEffect, createEvent, createStore } from "../store.ts";
import { setFaucetClient, setPublicClient, setSigner } from "./blockchain.ts";
import { $endpoint } from "./endpoint.ts";
import { setGlobalError } from "./error.ts";
import { $privateKey } from "./privateKey.ts";

// Stores
export const $smartAccount = createStore<SmartAccountV1 | null>(null);
export const $isSmartAccountInitialized = createStore(false);

// Events
export const setIsSmartAccountInitialized = createEvent<Boolean>();
export const setExistingSmartAccount = createEvent<SmartAccountV1>();
export const resetSmartAccount = createEvent();
export const retrySmartAccountCreation = createEvent();

// Effects
export const createSmartAccountFx = createEffect<
  { privateKey: Hex; endpoint: string },
  SmartAccountV1,
  Error
>(async ({ privateKey, endpoint }) => {
  try {
    const signer = await createSigner(privateKey);
    const client = await createClient(endpoint, 1);
    const faucetClient = await createFaucetClient(endpoint);

    const smartAccount = await initializeOrDeploySmartAccount({
      faucetClient: faucetClient,
      client: client,
      signer: signer,
      shardId: generateRandomShard(),
    });

    setSigner(signer);
    setPublicClient(client);
    setFaucetClient(faucetClient);

    await topUpAllTokens(smartAccount, faucetClient);

    await saveUserDetails({
      rpcEndpoint: endpoint,
      shardId: 1,
      privateKey: privateKey,
      smartAccountAddress: smartAccount.address,
    });

    return smartAccount;
  } catch (error) {
    console.error("Error during smartAccount creation:", error);
    setGlobalError("Failed to create smartAccount");
    throw new Error("Failed to create smartAccount");
  }
});

// Update smartAccount state on smartAccount creation success
$smartAccount.on(createSmartAccountFx.doneData, (_, smartAccount) => smartAccount);

// Update initialization state
$isSmartAccountInitialized.on(setIsSmartAccountInitialized, (_, isInitialized) => isInitialized);

// Set an existing smartAccount
$smartAccount.on(setExistingSmartAccount, (_, smartAccount) => smartAccount);

// Automatically create a smartAccount when conditions are met
sample({
  clock: combine($privateKey, $endpoint, $isSmartAccountInitialized, (pk, ep, isInitialized) => ({
    pk,
    ep,
    isInitialized,
  })),
  filter: ({ pk, ep, isInitialized }) => !isInitialized && pk !== null && ep.trim().length > 0,
  fn: ({ pk, ep }) => ({ privateKey: pk, endpoint: ep }),
  target: createSmartAccountFx,
});

// Retry smartAccount creation on the `retrySmartAccountCreation` event
sample({
  clock: retrySmartAccountCreation,
  source: combine($privateKey, $endpoint, (privateKey, endpoint) => ({ privateKey, endpoint })),
  filter: ({ privateKey, endpoint }) => privateKey !== null && endpoint.trim().length > 0,
  fn: ({ privateKey, endpoint }) => ({ privateKey, endpoint }),
  target: createSmartAccountFx,
});

// Log errors from smartAccount creation
createSmartAccountFx.failData.watch((error) => {
  console.error("Failed to create smartAccount:", error);
});

// Reset smartAccount and balances
$smartAccount.reset(resetSmartAccount);
$isSmartAccountInitialized.reset(resetSmartAccount);
