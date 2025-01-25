import {
  HttpTransport,
  LocalECDSAKeySigner,
  PublicClient,
  SmartAccountV1,
} from "@nilfoundation/niljs";
import type { Address } from "abitype";

export async function createClient(): Promise<{
  smartAccount: SmartAccountV1;
  publicClient: PublicClient;
  signer: LocalECDSAKeySigner;
}> {
  const smartAccountAddress = process.env.SMART_ACCOUNT_ADDR as
    | Address
    | undefined;

  if (!smartAccountAddress) {
    throw new Error("SMART_ACCOUNT_ADDR is not set in environment variables");
  }

  const endpoint = process.env.NIL_RPC_ENDPOINT;

  if (!endpoint) {
    throw new Error("NIL_RPC_ENDPOINT is not set in environment variables");
  }

  const publicClient = new PublicClient({
    transport: new HttpTransport({
      endpoint: endpoint,
    }),
    shardId: 1,
  });

  const signer = new LocalECDSAKeySigner({
    privateKey: `0x${process.env.PRIVATE_KEY}`,
  });
  const pubkey = await signer.getPublicKey();

  const smartAccount = new SmartAccountV1({
    pubkey: pubkey,
    address: smartAccountAddress,
    client: publicClient,
    signer,
  });

  return { smartAccount, publicClient, signer };
}
