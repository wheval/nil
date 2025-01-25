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

const anotherSmartAccount = await generateSmartAccount({
  shardId: 1,
  rpcEndpoint: RPC_ENDPOINT,
  faucetEndpoint: FAUCET_ENDPOINT,
});

console.log("smartAccountAddress", smartAccountAddress);

console.log("anotherSmartAccount", anotherSmartAccount.address);

const bounceAddress = generateRandomAddress();

// bounce transaction
const hash = await smartAccount.sendTransaction({
  to: anotherSmartAccount.address,
  value: 10_000_000n,
  bounceTo: bounceAddress,
  data: encodeFunctionData({
    abi: SmartAccountV1.abi,
    functionName: "syncCall",
    args: [smartAccountAddress, 100_000n, 10_000_000n, "0x"],
  }),
});

await waitTillCompleted(client, hash);

console.log("bounce address", bounceAddress);

const balance = await client.getBalance(bounceAddress, "latest");

console.log("balance", balance);

console.log("Transaction sent successfully");
