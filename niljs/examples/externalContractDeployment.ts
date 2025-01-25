import {
  HttpTransport,
  PublicClient,
  SmartAccountV1,
  bytesToHex,
  externalDeploymentTransaction,
  topUp,
} from "../src/index.js";
import { FAUCET_ENDPOINT, RPC_ENDPOINT, generatePublicKey } from "./helpers.js";

const client = new PublicClient({
  transport: new HttpTransport({
    endpoint: RPC_ENDPOINT,
  }),
  shardId: 1,
});

const pubKey = generatePublicKey();

const chainId = await client.chainId();
const gasPrice = await client.getGasPrice(1);

const deploymentTransaction = externalDeploymentTransaction(
  {
    salt: 100n,
    shard: 1,
    bytecode: SmartAccountV1.code,
    abi: SmartAccountV1.abi,
    args: [pubKey],
    feeCredit: 1_000_000n * gasPrice,
  },
  chainId,
);
const addr = bytesToHex(deploymentTransaction.to);

console.log("smartAccountAddress", addr);

await topUp({
  address: addr,
  faucetEndpoint: FAUCET_ENDPOINT,
  rpcEndpoint: RPC_ENDPOINT,
});

await deploymentTransaction.send(client);

while (true) {
  const code = await client.getCode(addr, "latest");
  if (code.length > 0) {
    console.log("code", bytesToHex(code));
    break;
  }
  await new Promise((resolve) => setTimeout(resolve, 1000));
}

console.log("Smart account deployed successfully");
