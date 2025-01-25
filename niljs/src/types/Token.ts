import type { Address } from "abitype";

/**
 * The structure representing a custom token.
 *
 * @export
 * @typedef {Token}
 */
export type Token = {
  id: Address;
  amount: bigint;
};
