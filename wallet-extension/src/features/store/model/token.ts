import type { SmartAccountV1 } from "@nilfoundation/niljs";
import { createEffect, createEvent, createStore, sample } from "effector";
import { persist } from "effector-storage/local";
import { fetchBalance, fetchSmartAccountCurrencies } from "../../blockchain";
import { btcAddress, ethAddress, usdtAddress } from "../../utils/currency.ts";
import { setGlobalError } from "./error.ts";
import { $smartAccount } from "./smartAccount.ts";

//Init
const initialTokens: { name: string; address: string; show: boolean; topupable: boolean }[] = [
  {
    address: "",
    name: "Nil",
    show: true,
    topupable: true,
  },
  {
    address: ethAddress,
    name: "ETH",
    show: true,
    topupable: true,
  },
  {
    address: usdtAddress,
    name: "USDT",
    show: true,
    topupable: true,
  },
  {
    address: btcAddress,
    name: "BTC",
    show: true,
    topupable: true,
  },
];

// Stores
export const $balance = createStore<bigint | null>(null);
export const $balanceCurrency = createStore<Record<string, bigint> | null>(null);
export const $tokens =
  createStore<{ name: string; address: string; show: boolean; topupable: boolean }[]>(
    initialTokens,
  );
persist({
  store: $tokens,
  key: "nil_wallet_tokens",
});

// Events
export const refetchBalancesEvent = createEvent();
export const hideToken = createEvent<string>();
export const showToken = createEvent<string>();
export const addToken = createEvent<{ name: string; address: string }>();

// Utils
export const getCurrencySymbolByAddress = (address: string): string => {
  return $tokens.getState().filter((token) => token.address === address)[0].name ?? "";
};

export const getBalanceForCurrency = (
  address: string,
  nilBalance: bigint,
  balanceCurrencies: Record<string, bigint>,
) => {
  if (address === "") {
    return nilBalance;
  }
  return balanceCurrencies[address] ?? 0n;
};

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

$tokens.on(addToken, (state, { name, address }) => [
  ...state,
  { name, address, show: true, topupable: false },
]);
$tokens.on(hideToken, (state, address) =>
  state.map((token) => (token.address === address ? { ...token, show: false } : token)),
);
$tokens.on(showToken, (state, address) =>
  state.map((token) => (token.address === address ? { ...token, show: true } : token)),
);

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
