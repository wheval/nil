import {
  type SmartAccountV1,
  getContract,
  waitTillCompleted,
} from "@nilfoundation/niljs";
import type { Address } from "abitype";

/**
 * Function to mint and send token from a contract.
 */
export async function mintAndSendToken({
  smartAccount,
  contractAddress,
  smartAccountAddress,
  mintAmount,
}: MintAndSendTokenArgs) {
  const TokenJson = require("../../artifacts/contracts/Token.sol/Token.json");
  const contract = getContract({
    abi: TokenJson.abi,
    address: contractAddress,
    client: smartAccount.client,
    smartAccount: smartAccount,
    externalInterface: {
      signer: smartAccount.signer,
      methods: ["mintToken", "sendToken"],
    },
  });

  const hash1 = await contract.external.mintToken([mintAmount]);
  await waitTillCompleted(smartAccount.client, hash1);
  const hash2 = await contract.external.sendToken([
    smartAccountAddress,
    contractAddress,
    mintAmount,
  ]);
  await waitTillCompleted(smartAccount.client, hash2);
}

export interface MintAndSendTokenArgs {
  smartAccount: SmartAccountV1;
  contractAddress: Address;
  smartAccountAddress: Address;
  mintAmount: bigint;
}
