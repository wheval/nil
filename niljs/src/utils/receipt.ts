import type { PublicClient } from "../clients/PublicClient.js";
import type { Hex } from "../types/Hex.js";
import type { ProcessedReceipt } from "../types/IReceipt.js";

/**
 * Makes it so that the client waits until the processing of the transaction whose hash is passed.
 *
 * @async
 * @param {PublicClient} client The client that must wait for action completion.
 * @param {Hex} hash The transaction hash.
 * @returns {unknown}
 * @example
 * await waitTillCompleted(client, hash);
 */
export async function waitTillCompleted(
  client: PublicClient,
  hash: Hex,
  options?: { waitTillMainShard?: boolean; interval?: number },
): Promise<ProcessedReceipt[]> {
  const interval = options?.interval || 1000;
  const waitTillMainShard = options?.waitTillMainShard || true;
  const receipts: ProcessedReceipt[] = [];
  const hashes: [Hex][] = [[hash]];
  let cur = 0;
  while (cur !== hashes.length) {
    const [hash] = hashes[cur];
    const receipt = await client.getTransactionReceiptByHash(hash);
    if (!receipt) {
      await new Promise((resolve) => setTimeout(resolve, interval));
      continue;
    }
    if (
      receipt.outTransactions !== null &&
      receipt.outputReceipts &&
      receipt.outputReceipts.filter((x) => x !== null).length !== receipt.outTransactions.length
    ) {
      await new Promise((resolve) => setTimeout(resolve, interval));
      continue;
    }
    if (waitTillMainShard && receipt.shardId !== 0 && !receipt.includedInMain) {
      await new Promise((resolve) => setTimeout(resolve, interval));
      continue;
    }
    cur++;
    receipts.push(receipt);
    if (receipt.outputReceipts) {
      for (const r of receipt.outputReceipts) {
        if (r !== null) hashes.push([r.transactionHash]);
      }
    }
  }

  return receipts;
}
