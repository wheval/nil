import {
  HttpTransport,
  LocalECDSAKeySigner,
  PublicClient,
  SmartAccountV1,
  topUp,
} from "@nilfoundation/niljs";
import type { Address } from "abitype";

export function createClient(
  url: string | undefined = undefined,
): PublicClient {
  if (!process.env.NIL_RPC_ENDPOINT) {
    throw new Error("NIL_RPC_ENDPOINT should not be null");
  }
  return new PublicClient({
    transport: new HttpTransport({
      endpoint: url || process.env.NIL_RPC_ENDPOINT,
    }),
    shardId: 1,
  });
}

export async function topUpSmartAccount(address: Address) {
  await topUp({
    address: address,
    faucetEndpoint:
      process.env.NIL_FAUCET_ENDPOINT || process.env.NIL_RPC_ENDPOINT,
    rpcEndpoint: process.env.NIL_RPC_ENDPOINT,
  });
}

export async function createSmartAccount(
  config: CreateSmartAccountConfig = {},
  client: PublicClient = createClient(),
): Promise<SmartAccountV1> {
  if (!process.env.PRIVATE_KEY) {
    throw new Error("PRIVATE_KEY should not be null");
  }
  const privateKey = process.env.PRIVATE_KEY.startsWith("0x")
    ? (process.env.PRIVATE_KEY as Address)
    : (`0x${process.env.PRIVATE_KEY}` as Address);
  const signer = new LocalECDSAKeySigner({
    privateKey: privateKey,
  });
  const smartAccountAddress =
    config.address ||
    (process.env.SMART_ACCOUNT_ADDR as `0x${string}` | undefined);

  const smartAccount = new SmartAccountV1({
    pubkey: signer.getPublicKey(),
    client: client,
    signer: signer,
    ...(smartAccountAddress
      ? { address: smartAccountAddress }
      : {
          salt: config.salt ?? BigInt(Math.round(Math.random() * 1000000)),
          shardId: config.shardId ?? 1,
        }),
  });

  if (config.faucetDeposit) {
    await topUpSmartAccount(smartAccount.address);

    const deployed = await smartAccount.checkDeploymentStatus();
    if (!deployed) {
      console.log("Deploying smart account", smartAccount.address);
      await smartAccount.selfDeploy();
    }
  }

  return smartAccount;
}

export type CreateSmartAccountConfig = {
  address?: Address | Uint8Array;
  salt?: Uint8Array | bigint;
  shardId?: number;
  faucetDeposit?: boolean;
};
