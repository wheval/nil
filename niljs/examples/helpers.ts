import {
  bytesToHex,
  calculateAddress,
  generateRandomPrivateKey,
  getPublicKey,
  refineSalt,
} from "../src/index.js";

export const RPC_ENDPOINT = "http://127.0.0.1:8529";
export const FAUCET_ENDPOINT = "http://127.0.0.1:8527";

export function generateRandomAddress(shardId = 1) {
  return bytesToHex(
    calculateAddress(
      shardId,
      new Uint8Array(1),
      refineSalt(BigInt(Math.floor(Math.random() * 10000))),
    ),
  );
}

export function generatePublicKey() {
  return getPublicKey(generateRandomPrivateKey());
}
