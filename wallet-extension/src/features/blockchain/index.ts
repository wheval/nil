export { fetchBalance, fetchSmartAccountCurrencies } from "./balance.ts";
export {
  topUpAllCurrencies,
  topUpSpecificCurrency,
  createFaucetClient,
} from "./faucet.ts";
export {
  createClient,
  createSigner,
  initializeOrDeploySmartAccount,
  sendCurrency,
} from "./smartAccount.ts";
