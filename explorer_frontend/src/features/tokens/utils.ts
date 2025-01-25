import { Token } from "./Token";
import btc from "./assets/btc.svg";
import eth from "./assets/eth.svg";
import nil from "./assets/nil.svg";
import usdt from "./assets/usdt.svg";

export const getTokenIcon = (name: string) => {
  switch (name) {
    case Token.ETH:
      return eth;
    case Token.NIL:
      return nil;
    case Token.USDT:
      return usdt;
    case Token.BTC:
      return btc;
    default:
      return null;
  }
};

export const ethAddress = "0x0001111111111111111111111111111111111112";
export const usdtAddress = "0x0001111111111111111111111111111111111113";
export const btcAddress = "0x0001111111111111111111111111111111111114";
export const nilAddress = "0x0001111111111111111111111111111111111110";

export const getTokenSymbolByAddress = (address: string) => {
  if (address === ethAddress) {
    return Token.ETH;
  }
  if (address === usdtAddress) {
    return Token.USDT;
  }
  if (address === btcAddress) {
    return Token.BTC;
  }
  return address;
};

export const getTokenAddressBySymbol = (symbol: string) => {
  if (symbol === Token.ETH) {
    return ethAddress;
  }

  if (symbol === Token.USDT) {
    return usdtAddress;
  }

  if (symbol === Token.BTC) {
    return btcAddress;
  }

  return symbol;
};
