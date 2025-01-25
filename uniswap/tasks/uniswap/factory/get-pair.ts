import { getContract } from "@nilfoundation/niljs";
import type { Address } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount } from "../../basic/basic";

task("get-pair", "Retrieve the pair address for the specified tokens")
  .addParam("factory", "The address of the Uniswap V2 factory")
  .addParam("token0", "The address of the first token")
  .addParam("token1", "The address of the second token")
  .setAction(async (taskArgs, _) => {
    // Destructure parameters for clarity
    const factoryAddress = taskArgs.factory as Address;
    const token0Address = taskArgs.token0 as Address;
    const token1Address = taskArgs.token1 as Address;

    const FactoryJson = require("../../artifacts/contracts/UniswapV2Factory.sol/UniswapV2Factory.json");
    const smartAccount = await createSmartAccount();
    const factory = getContract({
      abi: FactoryJson.abi,
      address: factoryAddress,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });

    // Retrieve the pair address
    const pairAddress = await factory.read.getTokenPair([
      token0Address,
      token1Address,
    ]);

    // Log the pair address
    console.log(`Pair address: ${pairAddress}`);
  });
