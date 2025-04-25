import { type IAddress, getContract, waitTillCompleted } from "@nilfoundation/niljs";
import type { HardhatRuntimeEnvironment } from "hardhat/types";
import type { Abi } from "viem";
import type { DeployContractConfig, GetContractAtConfig } from "../types";

export const getContractAt = async (
  { artifacts, network, nil }: HardhatRuntimeEnvironment,
  contractName: string,
  address: IAddress,
  config?: GetContractAtConfig,
) => {
  const [publicClient, smartAccount, contractArtifact] = await Promise.all([
    config?.publicClient ?? nil.getPublicClient(),
    config?.smartAccount ?? nil.getSmartAccount(),
    artifacts.readArtifact(contractName),
  ]);

  if (config?.signer) {
    return getContract({
      abi: contractArtifact.abi,
      address,
      client: publicClient,
      smartAccount: smartAccount,
      externalInterface: {
        signer: config.signer,
        methods: config?.externalMethods || [],
      },
    });
  }

  return getContract({
    abi: contractArtifact.abi,
    address,
    client: publicClient,
    smartAccount: smartAccount,
  });
};

export const deployContract = async (
  { artifacts, network, nil }: HardhatRuntimeEnvironment,
  contractName: string,
  args: unknown[] = [],
  config?: DeployContractConfig,
) => {
  const [publicClient, smartAccount, contractArtifact] = await Promise.all([
    config?.publicClient ?? nil.getPublicClient(),
    config?.smartAccount ?? nil.getSmartAccount(),
    artifacts.readArtifact(contractName),
  ]);

  console.log("Deploying contract", contractName, "with args", args);

  const { tx, address } = await smartAccount.deployContract({
    shardId: config?.shardId ?? smartAccount.shardId,
    args: args,
    bytecode: contractArtifact.bytecode as `0x${string}`,
    abi: contractArtifact.abi as Abi,
    salt: BigInt(Math.floor(Math.random() * 10000)),
    feeCredit: config?.feeCredit,
    value: config?.value,
  });
  await tx.wait();
  const receipts = await waitTillCompleted(publicClient, tx.hash);
  if (!receipts) {
    throw new Error("Transaction failed: receipts not found");
  }
  if (!receipts[0].success) {
    throw new Error(`Transaction failed: ${receipts[0].errorMessage}`);
  }
  console.log(`Deployed contract ${contractName} at address: ${address}, tx - ${tx.hash}`);

  const contract = getContract({
    abi: contractArtifact.abi,
    address,
    client: publicClient,
    smartAccount: smartAccount,
    externalInterface: {
      methods:
        config?.externalMethods ||
        contractArtifact.abi.filter((x) => x.onlyExternal === true).map((x) => x.name),
    },
  });
  return {
    address: address,
    ...contract,
  };
};
