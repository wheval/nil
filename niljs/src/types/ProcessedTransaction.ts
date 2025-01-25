import type { Address } from "abitype";
import type { Hex } from "./Hex.js";
import type { Flags } from "./RPCTransaction.js";

/**
 * The structure representing a processed transaction.
 *
 * @export
 * @typedef {ProcessedTransaction}
 */
export type ProcessedTransaction = {
  flags: Flags[];
  success: boolean;
  data: Hex;
  blockHash: Hex;
  blockNumber: number;
  from: Address;
  gasUsed: bigint;
  feeCredit: bigint;
  hash: Hex;
  seqno: bigint;
  to: Address;
  refundTo: Address;
  bounceTo: Address;
  index?: number;
  value: bigint;
  signature: Hex;
};
