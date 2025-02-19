import type { Abi } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount } from "../src/smart-account";
import { deployNilContract } from "../src/deploy";

task("deploy-incrementer").setAction(async (taskArgs, _) => {
  const smartAccount = await createSmartAccount();

  const IncrementerJson = require("../../artifacts/contracts/Incrementer.sol/Incrementer.json");

  const { contract, address } = await deployNilContract(
    smartAccount,
    IncrementerJson.abi as Abi,
    IncrementerJson.bytecode,
    [],
    smartAccount.shardId,
    [],
  );

  console.log("Incrementer contract deployed at address: " + address);

  await contract.write.increment([]);

  const value = await contract.read.getValue([]);

  console.log("Incrementer contract value: " + value);
});