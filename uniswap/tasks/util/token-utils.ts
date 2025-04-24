import { waitTillCompleted } from "@nilfoundation/niljs";
import type { Address } from "abitype";
import type { HardhatRuntimeEnvironment } from "hardhat/types";

/**
 * Function to mint and send token from a contract.
 */
export async function mintAndSendToken({
  hre,
  contractAddress,
  recipientAddress,
  mintAmount,
}: MintAndSendTokenArgs) {
  const contract = await hre.nil.getContractAt("Token", contractAddress, {});
  const client = await hre.nil.getPublicClient();

  const hash1 = await contract.write.mintTokenPublic([mintAmount]);
  await waitTillCompleted(client, hash1);
  const hash2 = await contract.write.sendTokenPublic([
    recipientAddress,
    contractAddress,
    mintAmount,
  ]);
  await waitTillCompleted(client, hash2, { waitTillMainShard: true });
  console.log("Sent token, tx:", hash2);
}

export interface MintAndSendTokenArgs {
  hre: HardhatRuntimeEnvironment;
  contractAddress: Address;
  recipientAddress: Address;
  mintAmount: bigint;
}
