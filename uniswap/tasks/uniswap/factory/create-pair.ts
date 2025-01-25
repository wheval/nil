import { getContract, waitTillCompleted } from "@nilfoundation/niljs";
import type { Address } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount } from "../../basic/basic";

task("create-pair", "Deploy and initialize a new Uniswap V2 pair")
  .addParam("factory", "The address of the Uniswap V2 factory")
  .addParam("token0", "The address of the first token")
  .addParam("token1", "The address of the second token")
  .setAction(async (taskArgs, _) => {
    // Destructure parameters for clarity
    const factoryAddress = taskArgs.factory as Address;
    const token0Address = taskArgs.token0 as Address;
    const token1Address = taskArgs.token1 as Address;
    const shardId = 1;

    const smartAccount = await createSmartAccount();

    const FactoryJson = require("../../artifacts/contracts/UniswapV2Factory.sol/UniswapV2Factory.json");
    const PairJson = require("../../artifacts/contracts/UniswapV2Pair.sol/UniswapV2Pair.json");

    const factory = getContract({
      abi: FactoryJson.abi,
      address: factoryAddress,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });

    // Create the pair and get its address
    const createPairTx = await factory.write.createPair([
      token0Address,
      token1Address,
      Math.floor(Math.random() * 10000000),
      shardId,
    ]);
    await waitTillCompleted(smartAccount.client, createPairTx);

    const pairAddress = await factory.read.getTokenPair([
      token0Address,
      token1Address,
    ]);

    // Log the pair address
    console.log(`Pair created successfully at address: ${pairAddress}`);

    // Attach to the newly created Uniswap V2 Pair contract
    const pair = getContract({
      abi: PairJson.abi,
      address: pairAddress as Address,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });

    // Initialize the pair with token addresses and IDs
    const hash = await pair.write.initialize([token0Address, token1Address]);
    await waitTillCompleted(smartAccount.client, hash);

    console.log(`Pair initialized successfully at address: ${pairAddress}`);
  });
