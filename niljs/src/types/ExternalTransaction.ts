/**
 * The external transaction type.
 *
 * @typedef {ExternalTransaction}
 */
type ExternalTransaction = {
  isDeploy: boolean;
  to: Uint8Array;
  chainId: number;
  seqno: number;
  data: Uint8Array;
  authData: Uint8Array;
  feeCredit?: bigint;
};

export type { ExternalTransaction };
