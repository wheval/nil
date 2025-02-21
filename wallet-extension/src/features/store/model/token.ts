import type { SmartAccountV1 } from "@nilfoundation/niljs";
import { createEffect, createEvent, createStore, sample } from "effector";
import { persist } from "effector-storage/local";
import { fetchBalance, fetchSmartAccountTokens } from "../../blockchain";
import { btcAddress, ethAddress, usdtAddress } from "../../utils/token.ts";
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
export const $balanceToken = createStore<Record<string, bigint> | null>(null);
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
export const getTokenSymbolByAddress = (address: string): string => {
  return $tokens.getState().filter((token) => token.address === address)[0].name ?? "";
};

export const getBalanceForToken = (
  address: string,
  nilBalance: bigint,
  balanceTokens: Record<string, bigint>,
) => {
  if (address === "") {
    return nilBalance;
  }
  return balanceTokens[address] ?? 0n;
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

export const fetchBalanceTokenssFx = createEffect<SmartAccountV1, Record<string, bigint>, Error>(
  async (smartAccount) => {
    try {
      const tokens = await fetchSmartAccountTokens(smartAccount);
      console.log("Fetched tokens:", tokens);

      return tokens;
    } catch (error) {
      console.error("Failed to fetch smartAccount tokens:", error);
      setGlobalError("Failed to fetch smartAccount tokens");
      throw error;
    }
  },
);

// Store updates
$balance.on(fetchBalanceFx.doneData, (_, balance) => balance);
$balanceToken.on(fetchBalanceTokenssFx.doneData, (_, tokens) => ({ ...tokens }));

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
  target: [fetchBalanceFx, fetchBalanceTokenssFx],
});

// Automatically fetch balances on smartAccount updates
sample({
  source: $smartAccount,
  clock: $smartAccount.updates,
  filter: (smartAccount) => smartAccount !== null && smartAccount !== undefined,
  fn: (smartAccount) => smartAccount as SmartAccountV1,
  target: [fetchBalanceFx, fetchBalanceTokenssFx],
});

// Watchers for debugging
$balance.watch((balance) => {
  console.log("Updated balance:", balance);
});

$balanceToken.watch((tokens) => {
  console.log("Updated balance tokens:", tokens);
});
