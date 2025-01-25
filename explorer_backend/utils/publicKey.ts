import {
  type RecoverPublicKeyParameters,
  type RecoverPublicKeyReturnType,
  concat,
  hexToNumber,
  isHex,
  stringToBytes,
  toHex,
} from "viem";

export const presignTransactionPrefix = "\x19Ethereum Signed Transaction:\n";

export async function recoverPublicKey({
  hash,
  signature,
}: RecoverPublicKeyParameters): Promise<RecoverPublicKeyReturnType> {
  const signatureHex = isHex(signature) ? signature : toHex(signature as Uint8Array);
  const hashHex = isHex(hash) ? hash : toHex(hash);

  // Derive v = recoveryId + 27 from end of the signature (27 is added when signing the transaction)
  // The recoveryId represents the y-coordinate on the secp256k1 elliptic curve and can have a value [0, 1].
  let v = hexToNumber(`0x${signatureHex.slice(130)}`);
  if (v === 0 || v === 1) v += 27;

  const { secp256k1 } = await import("@noble/curves/secp256k1");
  const publicKey = secp256k1.Signature.fromCompact(signatureHex.substring(2, 130))
    .addRecoveryBit(v - 27)
    .recoverPublicKey(hashHex.substring(2))
    .toHex(true);
  return `0x${publicKey}`;
}

export function prefixedTransaction(transaction: {
  raw: Buffer;
}): Uint8Array {
  const transactionBytes = (() => {
    return transaction.raw;
  })();
  const prefixBytes = stringToBytes(`${presignTransactionPrefix}${transactionBytes.length}`);
  return concat([prefixBytes, transactionBytes]);
}
