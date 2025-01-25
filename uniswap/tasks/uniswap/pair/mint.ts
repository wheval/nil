import { waitTillCompleted } from "@nilfoundation/niljs";
import { getContract } from "@nilfoundation/niljs";
import type { Address } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount, topUpSmartAccount } from "../../basic/basic";

task("mint", "Mint tokens and add liquidity to the pair")
  .addParam("pair", "The address of the pair contract")
  .addParam("amount0", "The amount of the first token to mint")
  .addParam("amount1", "The amount of the second token to mint")
  .setAction(async (taskArgs, _) => {
    const smartAccount = await createSmartAccount();

    // Destructure parameters for clarity
    const pairAddress = taskArgs.pair as Address;
    const amount0 = taskArgs.amount0 as number;
    const amount1 = taskArgs.amount1 as number;

    const PairJson = require("../../artifacts/contracts/UniswapV2Pair.sol/UniswapV2Pair.json");

    const pair = getContract({
      abi: PairJson.abi,
      address: pairAddress,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });

    // Fetch token addresses from the pair contract
    const token0Address = (await pair.read.token0Id([])) as Address;
    const token1Address = (await pair.read.token1Id([])) as Address;

    console.log("Token 0 Address:", token0Address);
    console.log("Token 1 Address:", token1Address);

    const TokenJson = require("../../artifacts/contracts/Token.sol/Token.json");
    // Attach to the Token contracts
    const token0 = getContract({
      abi: TokenJson.abi,
      address: token0Address,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });
    const token1 = getContract({
      abi: TokenJson.abi,
      address: token1Address,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });

    // Mint liquidity
    await topUpSmartAccount(pairAddress);
    console.log("Minting pair tokens...");

    const hash = await pair.write.mint([smartAccount.address], {
      tokens: [
        {
          id: token0Address,
          amount: BigInt(amount0),
        },
        {
          id: token1Address,
          amount: BigInt(amount1),
        },
      ],
    });

    await waitTillCompleted(smartAccount.client, hash);

    // Log balances in the pair contract
    const pairToken0Balance = await token0.read.getTokenBalanceOf([
      pairAddress,
    ]);
    console.log("Pair Balance 0:", pairToken0Balance);

    const pairToken1Balance = await token1.read.getTokenBalanceOf([
      pairAddress,
    ]);
    console.log("Pair Balance 1:", pairToken1Balance);

    console.log("Liquidity added...");

    // Retrieve and log reserves from the pair
    const [reserve0, reserve1] = (await pair.read.getReserves([])) as number[];
    console.log(
      `Reserves - Token0: ${reserve0.toString()}, Token1: ${reserve1.toString()}`,
    );

    // Check and log liquidity provider balance
    const lpBalance = (await pair.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as number;
    console.log(
      "Liquidity provider balance in smart account:",
      lpBalance.toString(),
    );

    // Retrieve and log total supply for the pair
    const totalSupply = (await pair.read.getTokenTotalSupply([])) as number;
    console.log("Total supply of pair tokens:", totalSupply.toString());
  });
