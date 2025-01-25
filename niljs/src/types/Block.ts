import type { Hex } from "./Hex.js";
import type { ProcessedTransaction } from "./ProcessedTransaction.js";

/**
 * The block type.
 * Type `T` determines whether the block contains processed transactions or just transaction hashes.
 */
type Block<T = false> = {
  number: number;
  hash: Hex;
  parentHash: Hex;
  inTransactionsRoot: Hex;
  receiptsRoot: Hex;
  shardId: number;
  transactions: T extends true ? Array<ProcessedTransaction> : Array<Hex>;
};

/**
 * The block tag type.
 */
type BlockTag = "latest" | "earliest" | "pending";

export type { Block, BlockTag };
