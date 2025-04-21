import { task } from "hardhat/config";
import { Wallet, ethers } from 'ethers';
import * as fs from "fs";
import {
  FaucetClient,
  HttpTransport,
  LocalECDSAKeySigner,
  PublicClient,
  SmartAccountV1,
  convertEthToWei,
  generateRandomPrivateKey,
  waitTillCompleted,
} from "@nilfoundation/niljs";
import "dotenv/config";
import { decodeFunctionResult, encodeFunctionData } from "viem";

let smartAccount: SmartAccountV1 | null = null;

async function getSmartAccount(): Promise<SmartAccountV1> {
  const rpcEndpoint = process.env.NIL_RPC_ENDPOINT as string;
  const client = new PublicClient({
    transport: new HttpTransport({ endpoint: rpcEndpoint }),
  });
  const faucetClient = new FaucetClient({
    transport: new HttpTransport({ endpoint: rpcEndpoint }),
  });

  const privateKey = process.env.NIL_PRIVATE_KEY as `0x${string}`;
  const smartAccountAddress = process.env.NIL_SMART_ACCOUNT_ADDRESS as `0x${string}`;

  if (privateKey && smartAccountAddress) {
    const signer = new LocalECDSAKeySigner({ privateKey });
    smartAccount = new SmartAccountV1({
      signer,
      client,
      address: smartAccountAddress,
      pubkey: signer.getPublicKey(),
    });
    console.log("🟢 Loaded Smart Account:", smartAccount.address);
  } else {
    const signer = new LocalECDSAKeySigner({ privateKey: privateKey });
    smartAccount = new SmartAccountV1({
      signer,
      client,
      salt: BigInt(Math.floor(Math.random() * 10000)),
      shardId: 1,
      pubkey: signer.getPublicKey(),
    });
    fs.writeFileSync("smartAccount.json", JSON.stringify({
      PRIVATE_KEY: privateKey,
      SMART_ACCOUNT_ADDRESS: smartAccount.address,
    }));
    console.log("🆕 New Smart Account Generated:", smartAccount.address);
  }

  const topUpFaucet = await faucetClient.topUp({
    smartAccountAddress: smartAccount.address,
    amount: ethers.parseEther("0.01"),
    faucetAddress: process.env.NIL as `0x${string}`,
  });

  await waitTillCompleted(client, topUpFaucet);

  if ((await smartAccount.checkDeploymentStatus()) === false) {
    await smartAccount.selfDeploy(true);
  }

  console.log("✅ Smart Account Funded (0.01 ETH)");

  // update 

  return smartAccount;
}

task("deploy-l2-eth-bridge-vault", "Deploys L2 ETHBridgeVault")
  .setAction(async () => {
    const deployerAccount = await getSmartAccount();
    if (!smartAccount) throw new Error("SmartAccount is not initialized.");
  });
