import type { ISignature } from "../signers/index.js";
import type { ITransaction } from "./ITransaction.js";

/**
 * The signed transaction interface.
 *
 * @typedef {ISignedTransaction}
 */
type ISignedTransaction = ITransaction & ISignature;

export type { ISignedTransaction };
