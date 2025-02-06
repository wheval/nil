import type { Hex } from "@nilfoundation/niljs";
import btc from "../../../public/icons/currency/btc.svg";
import custom from "../../../public/icons/currency/custom.svg";
import eth from "../../../public/icons/currency/ethereum.svg";
import nil from "../../../public/icons/currency/nil.svg";
import usdt from "../../../public/icons/currency/usdt.svg";
import { Currency } from "../components/currency";
import { getLast2Bytes } from "./address.ts";

// Converts a value in Wei (bigint) to Ether (string) with 18 decimal precision
export const convertWeiToEth = (wei: bigint, decimals = 18): string => {
  const eth = Number(wei) / 1e18;
  return Number.parseFloat(eth.toFixed(decimals)).toString();
};

// Retrieve faucet addresses from environment variables, throw an error if undefined
const ethAddress = import.meta.env.VITE_ETH_ADDRESS;
const usdtAddress = import.meta.env.VITE_USDT_ADDRESS;
const btcAddress = import.meta.env.VITE_BTC_ADDRESS;

if (!ethAddress) {
  throw new Error("Environment variable VITE_ETH_ADDRESS is not defined");
}

if (!usdtAddress) {
  throw new Error("Environment variable VITE_USDT_ADDRESS is not defined");
}

if (!btcAddress) {
  throw new Error("Environment variable VITE_BTC_ADDRESS is not defined");
}

// Returns the icon for the given currency name
export const getCurrencyIcon = (name: string) => {
  switch (name) {
    case Currency.ETH:
      return eth;
    case Currency.NIL:
      return nil;
    case Currency.USDT:
      return usdt;
    case Currency.BTC:
      return btc;
    default:
      return custom;
  }
};

// Returns the currency symbol (e.g., ETH, USDT) for a given token address
export const getCurrencySymbolByAddress = (address: string): string => {
  if (address === ethAddress) {
    return Currency.ETH;
  }
  if (address === usdtAddress) {
    return Currency.USDT;
  }
  if (address === btcAddress) {
    return Currency.BTC;
  }
  return "";
};

// Returns the token address (e.g., 0x...) for a given currency symbol
export const getTokenAddressBySymbol = (symbol: string): string => {
  if (symbol === Currency.ETH) {
    return ethAddress;
  }

  if (symbol === Currency.USDT) {
    return usdtAddress;
  }

  if (symbol === Currency.BTC) {
    return btcAddress;
  }

  return "";
};

export const getTokenAddress = (
  symbol: string,
  balanceCurrencies: Record<string, bigint> | null,
): string => {
  let tokenAddress = getTokenAddressBySymbol(symbol);

  // If token address is empty, find it from balanceCurrencies
  if (!tokenAddress && balanceCurrencies) {
    for (const address of Object.keys(balanceCurrencies)) {
      let detectedSymbol = getCurrencySymbolByAddress(address);
      if (!detectedSymbol) {
        detectedSymbol = getLast2Bytes(address as Hex);
      }
      if (detectedSymbol === symbol) {
        tokenAddress = address;
        break;
      }
    }
  }

  return tokenAddress;
};

export function getCurrencies(balanceCurrencies: Record<string, bigint> | null) {
  const dynamicCurrencies = balanceCurrencies
    ? (Object.keys(balanceCurrencies)
        .map((address) => {
          let symbol = getCurrencySymbolByAddress(address);
          if (!symbol) {
            symbol = getLast2Bytes(address as Hex);
          }

          const icon = getCurrencyIcon(symbol);
          return icon ? { icon, label: symbol } : null;
        })
        .filter(Boolean) as { icon: string; label: string }[])
    : [];

  return [{ icon: "/icons/currency/nil.svg", label: Currency.NIL }, ...dynamicCurrencies];
}

// Get the correct balance
export function getBalanceForCurrency(
  selectedCurrency: string,
  nilBalance: bigint | null,
  balanceCurrencies: Record<string, bigint> | null,
): bigint {
  return selectedCurrency === Currency.NIL
    ? nilBalance || BigInt(0)
    : balanceCurrencies?.[getTokenAddress(selectedCurrency, balanceCurrencies)] || BigInt(0);
}
