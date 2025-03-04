import type { SmartAccountV1 } from "@nilfoundation/niljs";
import { createEffect, createEvent, createStore, sample } from "effector";
import { type Token, getTokens, saveToken, setTokens } from "../../../background/storage/tokens.ts";
import { fetchBalance, fetchSmartAccountTokens } from "../../blockchain";
import { TokenNames } from "../../components/token";
import { btcAddress, ethAddress, usdtAddress } from "../../utils/token.ts";
import { setGlobalError } from "./error.ts";
import { $smartAccount } from "./smartAccount.ts";

//Init
export const initialTokens: Token[] = [
  {
    address: "",
    name: TokenNames.NIL,
    show: true,
    topupable: true,
  },
  {
    address: ethAddress,
    name: TokenNames.ETH,
    show: true,
    topupable: true,
  },
  {
    address: usdtAddress,
    name: TokenNames.USDT,
    show: true,
    topupable: true,
  },
  {
    address: btcAddress,
    name: TokenNames.BTC,
    show: true,
    topupable: true,
  },
];

// Stores
export const $balance = createStore<bigint | null>(null);
export const $balanceToken = createStore<Record<string, bigint> | null>(null);
export const $tokens = createStore<Token[]>(initialTokens);

// Events
export const refetchBalancesEvent = createEvent();
export const initializeTokens = createEvent<string>();
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

export const addTokensFx = createEffect<Token, void, Error>(async (token) => {
  try {
    await saveToken(token);
  } catch (error) {
    console.error("Failed to add token:", error);
    throw error;
  }
});

export const changeTokensFx = createEffect<{ address: string; show: boolean }, void, Error>(
  async (token) => {
    try {
      await setTokens(
        $tokens
          .getState()
          .map((t) => (t.address === token.address ? { ...t, show: token.show } : t)),
      );
    } catch (error) {
      console.error("Failed to change token:", error);
      throw error;
    }
  },
);

export const fetchTokensFx = createEffect<string, Token[], Error>(async (_) => {
  try {
    const tokens = await getTokens();
    if (tokens.length === 0) {
      await setTokens(initialTokens);
      return initialTokens;
    }

    return tokens;
  } catch (error) {
    console.error("Failed to fetch tokens:", error);
    throw error;
  }
});

export const fetchBalanceTokensFx = createEffect<SmartAccountV1, Record<string, bigint>, Error>(
  async (smartAccount) => {
    try {
      return await fetchSmartAccountTokens(smartAccount);
    } catch (error) {
      console.error("Failed to fetch smartAccount tokens:", error);
      setGlobalError("Failed to fetch smartAccount tokens");
      throw error;
    }
  },
);

// Store updates
$balance.on(fetchBalanceFx.doneData, (_, balance) => balance);
$balanceToken.on(fetchBalanceTokensFx.doneData, (_, tokens) => ({ ...tokens }));

$tokens.on(fetchTokensFx.doneData, (_, tokens) => tokens);

sample({
  source: addToken,
  fn: (token) => ({ ...token, show: true, topupable: false }),
  target: addTokensFx,
});

sample({
  source: hideToken,
  fn: (token) => ({ address: token, show: false }),
  target: changeTokensFx,
});

sample({
  source: showToken,
  fn: (token) => ({ address: token, show: true }),
  target: changeTokensFx,
});

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
  target: [fetchBalanceFx, fetchBalanceTokensFx],
});

// Automatically fetch balances on smartAccount updates
sample({
  source: $smartAccount,
  clock: $smartAccount.updates,
  filter: (smartAccount) => smartAccount !== null && smartAccount !== undefined,
  fn: (smartAccount) => smartAccount as SmartAccountV1,
  target: [fetchBalanceFx, fetchBalanceTokensFx],
});

sample({
  source: initializeTokens,
  fn: (str) => str,
  target: fetchTokensFx,
});

// Watchers for debugging
$balance.watch((balance) => {
  console.log("Updated balance:", balance);
});

$balanceToken.watch((tokens) => {
  console.log("Updated balance tokens:", tokens);
});
