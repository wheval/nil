import type { Hex } from "./Hex.js";
import type { ILog } from "./ILog.js";
import type { Flags } from "./RPCTransaction.js";

/**
 * The receipt interface.
 */
type IReceipt = {
  flags: Flags[];
  success: boolean;
  gasUsed: string;
  gasPrice?: string;
  bloom: string;
  logs: ILog[];
  transactionHash: Hex;
  contractAddress: string;
  blockHash: string;
  blockNumber: number;
  txnIndex: number;
  outTransactions: Hex[] | null;
  outputReceipts: (IReceipt | null)[] | null;
  shardId: number;
  includedInMain: boolean;
};

type ProcessedReceipt = Omit<IReceipt, "gasUsed" | "gasPrice" | "outputReceipts"> & {
  gasUsed: bigint;
  gasPrice?: bigint;
  outputReceipts: (ProcessedReceipt | null)[] | null;
};

export type { IReceipt, ProcessedReceipt };
