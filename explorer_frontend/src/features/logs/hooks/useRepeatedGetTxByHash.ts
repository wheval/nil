import type { Hex } from "@nilfoundation/niljs";
import { useEffect, useState } from "react";
import { fetchTransactionByHash } from "../../../api/transaction";
import type { Transaction } from "../../transaction/types/Transaction";

export const useRepeatedGetTxByHash = (
  txHash: Hex,
  interval = 1000,
  retries = 3,
): {
  data: Transaction | null;
  error: boolean;
  loading: boolean;
} => {
  const [data, setData] = useState<Transaction | null>(null);
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!txHash) {
      return;
    }

    const controller = new AbortController();
    const signal = controller.signal;
    let retriesCount = 0;

    const fetchData = async () => {
      try {
        const response = await fetchTransactionByHash(txHash, {
          signal,
        });

        if (response) {
          setData(response);
          setError(false);
          setLoading(false);
          clearInterval(intervalId);
        }
      } catch (err) {
        // biome-ignore lint/suspicious/noExplicitAny: <explanation>
        if ((err as any).name === "AbortError") {
          return;
        }

        if (retriesCount >= retries) {
          setError(true);
          setLoading(false);
          clearInterval(intervalId);
          return;
        }
      } finally {
        retriesCount++;
      }
    };

    const intervalId = setInterval(fetchData, interval);

    return () => {
      clearInterval(intervalId);
      controller.abort();
    };
  }, [interval, txHash, retries]);

  return { data, error, loading };
};
