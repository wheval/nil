import { bytesToHex, hexToBytes, numberToBytesBE } from "@noble/curves/abstract/utils";
import { secp256k1 } from "@noble/curves/secp256k1";
import { keccak_256 } from "@noble/hashes/sha3";
import { toHex } from "../encoding/toHex.js";
import type { Hex } from "../types/Hex.js";
import { assertIsValidShardId } from "../utils/assert.js";
import { addHexPrefix, removeHexPrefix } from "../utils/hex.js";
import type { IAddress } from "./types/IAddress.js";
import type { IPrivateKey } from "./types/IPrivateKey.js";

/**
 * Returns the public key from the private key using the secp256k1 curve.
 */
const getPublicKey = (privateKey: IPrivateKey, isCompressed = false): Hex => {
  const publicKey = secp256k1.getPublicKey(removeHexPrefix(privateKey), isCompressed);

  return addHexPrefix(bytesToHex(publicKey));
};

/**
 * Returns the address from the public key.
 * @param publicKey The public key in hex format
 * @param shardId The ID of the shard where the address is located.
 * @returns The address in hex format.
 */
const getAddressFromPublicKey = (publicKey: Hex, shardId: number): IAddress => {
  assertIsValidShardId(shardId);
  const publickKeyWithoutPrefix = removeHexPrefix(publicKey);

  const pubKeyBytes =
    typeof publickKeyWithoutPrefix === "string"
      ? hexToBytes(publickKeyWithoutPrefix)
      : publickKeyWithoutPrefix;

  const hashPart = keccak_256(pubKeyBytes);
  const shardPart = numberToBytesBE(shardId, 2);
  return toHex(new Uint8Array([...shardPart, ...hashPart.slice(14)]));
};

export { getPublicKey, getAddressFromPublicKey };
