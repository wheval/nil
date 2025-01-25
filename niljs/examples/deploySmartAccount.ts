import {
  HttpTransport,
  LocalECDSAKeySigner,
  PublicClient,
  SmartAccountV1,
  bytesToHex,
  generateRandomPrivateKey,
  topUp,
} from "../src/index.js";
import { FAUCET_ENDPOINT, RPC_ENDPOINT } from "./helpers.js";

const client = new PublicClient({
  transport: new HttpTransport({
    endpoint: RPC_ENDPOINT,
  }),
  shardId: 1,
});

const signer = new LocalECDSAKeySigner({
  privateKey: generateRandomPrivateKey(),
});

const pubkey = signer.getPublicKey();

const smartAccount = new SmartAccountV1({
  pubkey: pubkey,
  salt: 100n,
  shardId: 1,
  client,
  signer,
});
const smartAccountAddress = smartAccount.address;

console.log("smartAccountAddress", smartAccountAddress);

await topUp({
  address: smartAccountAddress,
  faucetEndpoint: FAUCET_ENDPOINT,
  rpcEndpoint: RPC_ENDPOINT,
});

await smartAccount.selfDeploy(true);

const code = await client.getCode(smartAccountAddress, "latest");

console.log("code", bytesToHex(code));

console.log("Smart account deployed successfully");
