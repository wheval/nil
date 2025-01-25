import type { Hex } from "@nilfoundation/niljs";
import { createDomain } from "effector";

export const faucetsDomain = createDomain("faucet");

export const $faucets = faucetsDomain.createStore<Record<string, Hex> | null>(null);

export const fetchFaucetsEvent = faucetsDomain.createEvent();

export const fetchFaucetsFx = faucetsDomain.createEffect<string, Record<string, Hex>>();
