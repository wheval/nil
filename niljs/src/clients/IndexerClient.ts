import type { IAddress } from "../signers/types/IAddress.js";
import type { Hex } from "../types/Hex.js";
import { BaseClient } from "./BaseClient.js";

export class IndexerClient extends BaseClient {
  /**
   * Gets address actions page
   * @param address - The address to get actions for.
   * @param sinceTimestamp - The timestamp to get actions since.
   * @returns The page of address actions.
   */
  public async getAddressActions(address: Hex, sinceTimestamp = 0) {
    return await this.request<AddressAction[]>({
      method: "indexer_getAddressActions",
      params: [address, sinceTimestamp],
    });
  }
}

export type AddressAction = {
  hash: Hex;
  from: IAddress;
  to: IAddress;
  amount: bigint;
  timestamp: number;
  blockId: number;
  type: AddressActionKind;
  status: AddressActionStatus;
};

export enum AddressActionKind {
  SendEth = "SendEth",
  ReceiveEth = "ReceiveEth",
  SmartContractCall = "SmartContractCall",
}

export enum AddressActionStatus {
  Success = "Success",
  Failed = "Failed",
}
