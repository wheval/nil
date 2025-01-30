import { HttpTransport, PublicClient, generateSmartAccount } from "../src/index.js";
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

console.log("smartAccountAddress", smartAccountAddress);

const anotherAddress = generateRandomAddress();

await smartAccount.syncSendTransaction({
  to: anotherAddress,
  value: 10n,
  gas: 100_000n * 10n,
  maxPriorityFeePerGas: 10n,
  maxFeePerGas: 1_000_000_000_000n,
});

while (true) {
  const balance = await client.getBalance(anotherAddress, "latest");
  if (balance > 0) {
    console.log("balance", balance);
    break;
  }
  await new Promise((resolve) => setTimeout(resolve, 1000));
}

console.log("Transaction sent successfully");
