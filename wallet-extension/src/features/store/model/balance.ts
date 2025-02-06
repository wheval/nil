import type { SmartAccountV1 } from "@nilfoundation/niljs";
import { createEffect, createEvent, createStore, sample } from "effector";
import { fetchBalance, fetchSmartAccountCurrencies } from "../../blockchain";
import { setGlobalError } from "./error.ts";
import { $smartAccount } from "./smartAccount.ts";

// Stores
export const $balance = createStore<bigint | null>(null);
export const $balanceCurrency = createStore<Record<string, bigint> | null>(null);

// Events
export const refetchBalancesEvent = createEvent();

// Effects
export const fetchBalanceFx = createEffect<SmartAccountV1, bigint, Error>(async (smartAccount) => {
  try {
    return await fetchBalance(smartAccount);
  } catch (error) {
    console.error("Failed to fetch balance:", error);
    setGlobalError("Failed to fetch balance");
    throw error;
  }
});

export const fetchBalanceCurrenciesFx = createEffect<SmartAccountV1, Record<string, bigint>, Error>(
  async (smartAccount) => {
    try {
      const currencies = await fetchSmartAccountCurrencies(smartAccount);
      console.log("Fetched currencies:", currencies);

      return currencies;
    } catch (error) {
      console.error("Failed to fetch smartAccount currencies:", error);
      setGlobalError("Failed to fetch smartAccount currencies");
      throw error;
    }
  },
);

// Store updates
$balance.on(fetchBalanceFx.doneData, (_, balance) => balance);
$balanceCurrency.on(fetchBalanceCurrenciesFx.doneData, (_, currencies) => ({ ...currencies }));

// Automatically fetch balances when `refetchBalancesEvent` is triggered
sample({
  source: $smartAccount,
  clock: refetchBalancesEvent,
  filter: (smartAccount) => smartAccount !== null && smartAccount !== undefined,
  fn: (smartAccount) => smartAccount as SmartAccountV1,
  target: [fetchBalanceFx, fetchBalanceCurrenciesFx],
});

// Automatically fetch balances on smartAccount updates
sample({
  source: $smartAccount,
  clock: $smartAccount.updates,
  filter: (smartAccount) => smartAccount !== null && smartAccount !== undefined,
  fn: (smartAccount) => smartAccount as SmartAccountV1,
  target: [fetchBalanceFx, fetchBalanceCurrenciesFx],
});

// Watchers for debugging
$balance.watch((balance) => {
  console.log("Updated balance:", balance);
});

$balanceCurrency.watch((currencies) => {
  console.log("Updated balance currencies:", currencies);
});
