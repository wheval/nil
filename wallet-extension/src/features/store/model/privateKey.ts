import type { Hex } from "@nilfoundation/niljs";
import { generateRandomPrivateKey } from "@nilfoundation/niljs";
import { createEvent, createStore } from "../store.ts";

// Private Key
export const setPrivateKey = createEvent<Hex>();
export const initializePrivateKey = createEvent();

export const $privateKey = createStore<Hex | null>(null)
  // Update the private key when `setPrivateKey` is triggered
  .on(setPrivateKey, (_, privateKey) => privateKey)

  // Generate a private key only if it doesn't already exist
  .on(initializePrivateKey, (state) => {
    if (state === null) {
      return generateRandomPrivateKey();
    }
    return state;
  });
