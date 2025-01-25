import type { Abi } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount } from "../../basic/basic";
import { deployNilContract } from "../../util/deploy";

task("deploy-factory").setAction(async (taskArgs, _) => {
  const smartAccount = await createSmartAccount();

  const FactoryJson = require("../../../artifacts/contracts/UniswapV2Factory.sol/UniswapV2Factory.json");

  const { contract, address } = await deployNilContract(
    smartAccount,
    FactoryJson.abi as Abi,
    FactoryJson.bytecode,
    [smartAccount.address],
    smartAccount.shardId,
    [],
  );
  console.log("Uniswap factory contract deployed at address: " + address);
});
