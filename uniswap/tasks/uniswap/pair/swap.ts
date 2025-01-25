import { waitTillCompleted } from "@nilfoundation/niljs";
import { getContract } from "@nilfoundation/niljs";
import type { Address } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount } from "../../basic/basic";

task("swap", "Swap token0 for token1 in the Uniswap pair")
  .addParam("pair", "The address of the Uniswap pair contract")
  .addParam("amount", "The amount of token0 to swap")
  .setAction(async (taskArgs, _) => {
    const smartAccount = await createSmartAccount();

    // Destructure parameters for clarity
    const pairAddress = taskArgs.pair as Address;
    const swapAmount = BigInt(taskArgs.amount);

    const TokenJson = require("../../artifacts/contracts/Token.sol/Token.json");
    const PairJson = require("../../artifacts/contracts/UniswapV2Pair.sol/UniswapV2Pair.json");

    const pair = getContract({
      abi: PairJson.abi,
      address: pairAddress,
      client: smartAccount.client,
      smartAccount: smartAccount,
    });

    // Retrieve token addresses from the pair contract
    const token0Address = (await pair.read.token0Id([])) as Address;
    const token1Address = (await pair.read.token1Id([])) as Address;

    console.log("Token 0 Address:", token0Address);
    console.log("Token 1 Address:", token1Address);

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

    // Retrieve reserves from the pair
    const reserves = await pair.read.getReserves([]);
    // @ts-ignore
    const reserve0 = reserves[0] as bigint;
    // @ts-ignore
    const reserve1 = reserves[1] as bigint;
    console.log(`Reserves - Token0: ${reserve0}, Token1: ${reserve1}`);

    // Calculate expected output amount for the swap
    const expectedOutputAmount = calculateOutputAmount(
      swapAmount,
      reserve0,
      reserve1,
    );
    console.log(
      "Expected output amount for swap:",
      expectedOutputAmount.toString(),
    );

    // Log balances before the swap
    const balanceToken0Before = (await token0.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    const balanceToken1Before = (await token1.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log(
      "Balance of token0 before swap:",
      balanceToken0Before.toString(),
    );
    console.log(
      "Balance of token1 before swap:",
      balanceToken1Before.toString(),
    );

    // Execute the swap
    console.log("Executing swap...");

    const hash = await pair.write.swap(
      [0, expectedOutputAmount, smartAccount.address],
      {
        tokens: [
          {
            id: token0Address,
            amount: swapAmount,
          },
        ],
      },
    );

    await waitTillCompleted(smartAccount.client, hash);

    console.log("Swap executed successfully.");

    // Log balances after the swap
    const balanceToken0After = (await token0.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    const balanceToken1After = (await token1.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log("Balance of token0 after swap:", balanceToken0After.toString());
    console.log("Balance of token1 after swap:", balanceToken1After.toString());
  });

// Function to calculate the output amount for the swap
function calculateOutputAmount(
  amountIn: bigint,
  reserveIn: bigint,
  reserveOut: bigint,
): bigint {
  const amountInWithFee = amountIn * BigInt(997);
  const numerator = amountInWithFee * reserveOut;
  const denominator = reserveIn * BigInt(1000) + amountInWithFee;
  return numerator / denominator;
}
