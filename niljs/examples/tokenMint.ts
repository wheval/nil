import { encodeFunctionData } from "viem";
import {
  HttpTransport,
  PublicClient,
  SmartAccountV1,
  generateSmartAccount,
  waitTillCompleted,
} from "../src/index.js";
import { FAUCET_ENDPOINT, RPC_ENDPOINT, generateRandomAddress } from "./helpers.js";

const client = new PublicClient({
  transport: new HttpTransport({
    endpoint: RPC_ENDPOINT,
  }),
  shardId: 1,
});

const smartAccount = await generateSmartAccount({
  shardId: 1,
  rpcEndpoint: RPC_ENDPOINT,
  faucetEndpoint: FAUCET_ENDPOINT,
});
const smartAccountAddress = smartAccount.address;

console.log("Smart account deployed successfully");
console.log("smartAccountAddress", smartAccountAddress);

const gasPrice = await client.getGasPrice(1);

const hashTransaction = await smartAccount.sendTransaction({
  to: smartAccountAddress,
  feeCredit: 1_000_000n * gasPrice,
  value: 0n,
  data: encodeFunctionData({
    abi: SmartAccountV1.abi,
    functionName: "setTokenName",
    args: ["MY_TOKEN"],
  }),
});

await waitTillCompleted(client, hashTransaction);

const hashTransaction2 = await smartAccount.sendTransaction({
  to: smartAccountAddress,
  feeCredit: 1_000_000n * gasPrice,
  value: 0n,
  data: encodeFunctionData({
    abi: SmartAccountV1.abi,
    functionName: "mintToken",
    args: [100_000_000n],
  }),
});

await waitTillCompleted(client, hashTransaction2);

const tokens = await client.getTokens(smartAccountAddress, "latest");

console.log("tokens", tokens);

const anotherAddress = generateRandomAddress();

const sendHash = await smartAccount.sendTransaction({
  to: anotherAddress,
  value: 10_000_000n,
  feeCredit: 100_000n * gasPrice,
  tokens: [
    {
      id: smartAccountAddress,
      amount: 100_00n,
    },
  ],
});

await waitTillCompleted(client, sendHash);

const anotherTokens = await client.getTokens(anotherAddress, "latest");

console.log("anotherTokens", anotherTokens);
