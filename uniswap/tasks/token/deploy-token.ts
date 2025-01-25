import { bytesToHex } from "@nilfoundation/niljs";
import type { Abi } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount, topUpSmartAccount } from "../basic/basic";
import { deployNilContract } from "../util/deploy";

task("deploy-token")
  .addParam("amount")
  .setAction(async (taskArgs, _) => {
    const smartAccount = await createSmartAccount({ faucetDeposit: true });

    const TokenJson = require("../../artifacts/contracts/Token.sol/Token.json");
    const { contract, address } = await deployNilContract(
      smartAccount,
      TokenJson.abi as Abi,
      TokenJson.bytecode,
      ["Token", bytesToHex(smartAccount.signer.getPublicKey())],
      smartAccount.shardId,
      ["mintToken"],
    );
    console.log("Token contract deployed at address: " + address);

    await topUpSmartAccount(address);

    // @ts-ignore
    const hash = await contract.external.mintToken([taskArgs.amount]);
    console.log("Minted token with hash: " + hash);
  });
