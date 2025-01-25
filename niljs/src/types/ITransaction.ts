import type { ValueOf } from "@chainsafe/ssz";
import type { SszTransactionSchema } from "../encoding/ssz.js";

/**
 * The interface for the transaction object. This object is used to represent a transaction in the client code.
 * It may differ from the actual transaction object used inside the network.
 */
interface ITransaction extends ValueOf<typeof SszTransactionSchema> {}

export type { ITransaction };
