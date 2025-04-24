import { topUpSmartAccount } from "@nilfoundation/hardhat-nil-plugin/src";
import { waitTillCompleted } from "@nilfoundation/niljs";
import { task } from "hardhat/config";
import type { HttpNetworkConfig } from "hardhat/src/types/config";
import { calculateOutputAmount } from "../util/math";
import { mintAndSendToken } from "../util/token-utils";

task("demo-router", "Run demo with Uniswap Router").setAction(
  async (taskArgs, hre) => {
    const shardId = 1;
    const mintAmount = BigInt(100000);
    const mintToken0Amount = 10000;
    const mintToken1Amount = 10000;
    const swapAmount = 1000;

    const smartAccount = await hre.nil.getSmartAccount();
    const client = await hre.nil.getPublicClient();
    const rpc = (hre.network.config as HttpNetworkConfig).url;

    const factory = await hre.nil.deployContract("UniswapV2Factory", [
      smartAccount.address,
    ]);
    const token0 = await hre.nil.deployContract("Token", [
      "Token0",
      100000000000n,
    ]);
    await topUpSmartAccount(token0.address, rpc);
    const token1 = await hre.nil.deployContract("Token", [
      "Token0",
      100000000000n,
    ]);
    await topUpSmartAccount(token1.address, rpc);

    console.log("Factory deployed " + factory.address);
    console.log("Token0 deployed " + token0.address);
    console.log("Token1 deployed " + token1.address);

    const router = await hre.nil.deployContract("UniswapV2Router01", []);
    console.log("Router deployed " + router.address);

    // 1. CREATE PAIR
    const pairTxHash = await factory.write.createPair([
      token0.address,
      token1.address,
      Math.floor(Math.random() * 10000000),
      shardId,
    ]);

    await waitTillCompleted(client, pairTxHash);

    const pairAddress = (await factory.read.getTokenPair([
      token0.address,
      token1.address,
    ])) as `0x${string}`;

    // Log the pair address
    console.log(`Pair created successfully at address: ${pairAddress}`);

    const pair = await hre.nil.getContractAt("UniswapV2Pair", pairAddress);

    // Initialize the pair with token addresses and IDs
    await pair.write.initialize([token0.address, token1.address]);

    console.log(`Pair initialized successfully at address: ${pairAddress}`);

    // 2. MINT TOKENS
    console.log(
      `Minting ${mintAmount} Token0 to smart account ${smartAccount.address}...`,
    );
    await mintAndSendToken({
      hre,
      contractAddress: token0.address,
      recipientAddress: smartAccount.address,
      mintAmount,
    });

    // Mint and send Token1
    console.log(
      `Minting ${mintAmount} Token1 to smart account ${smartAccount.address}...`,
    );
    await mintAndSendToken({
      hre,
      contractAddress: token1.address,
      recipientAddress: smartAccount.address,
      mintAmount,
    });

    // Verify the balance of the recipient smart account for both tokens
    const recipientBalanceToken0 = await token0.read.getTokenBalanceOf([
      smartAccount.address,
    ]);
    const recipientBalanceToken1 = await token1.read.getTokenBalanceOf([
      smartAccount.address,
    ]);

    console.log(
      `Recipient balance after transfer - Token0: ${recipientBalanceToken0}, Token1: ${recipientBalanceToken1}`,
    );

    // 3. ROUTER: ADD LIQUIDITY

    // Mint liquidity
    console.log("Adding liquidity...");

    const hash = await router.write.addLiquidity(
      [pairAddress, smartAccount.address],
      {
        tokens: [
          {
            id: token0.address,
            amount: BigInt(mintToken0Amount),
          },
          {
            id: token1.address,
            amount: BigInt(mintToken1Amount),
          },
        ],
      },
    );

    await waitTillCompleted(client, hash);

    // Log balances in the pair contract
    const pairToken0Balance = (await token0.read.getTokenBalanceOf([
      pairAddress,
    ])) as number;
    console.log("Pair Balance of Token0:", pairToken0Balance);

    const pairToken1Balance = (await token1.read.getTokenBalanceOf([
      pairAddress,
    ])) as number;
    console.log("Pair Balance of Token1:", pairToken1Balance);

    console.log("Liquidity added...");

    // Retrieve and log reserves from the pair
    // @ts-ignore
    const [reserve0, reserve1] = await pair.read.getReserves([]);
    console.log(
      `ADDLIQUIDITY RESULT: Reserves - Token0: ${reserve0.toString()}, Token1: ${reserve1.toString()}`,
    );

    // Check and log liquidity provider balance
    const lpBalance = (await pair.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log(
      "ADDLIQUIDITY RESULT: Liquidity provider balance in smart account:",
      lpBalance.toString(),
    );

    // Retrieve and log total supply for the pair
    const totalSupply = (await pair.read.getTokenTotalSupply([])) as bigint;
    console.log(
      "ADDLIQUIDITY RESULT: Total supply of pair tokens:",
      totalSupply.toString(),
    );

    // 4. ROUTER: SWAP
    const expectedOutputAmount = calculateOutputAmount(
      BigInt(swapAmount),
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

    // Send token0 to the pair contract
    const hash2 = await router.write.swap(
      [smartAccount.address, pairAddress, 0, expectedOutputAmount],
      {
        tokens: [
          {
            id: token0.address,
            amount: BigInt(swapAmount),
          },
        ],
      },
    );

    await waitTillCompleted(client, hash2);

    console.log(
      `Sent ${swapAmount.toString()} of token0 to the pair contract. Tx - ${hash2}`,
    );

    console.log("Swap executed successfully.");

    // Log balances after the swap
    const balanceToken0After = (await token0.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    const balanceToken1After = (await token1.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log(
      "SWAP RESULT: Balance of token0 after swap:",
      balanceToken0After.toString(),
    );
    console.log(
      "SWAP RESULT: Balance of token1 after swap:",
      balanceToken1After.toString(),
    );

    // 5. ROUTER: REMOVE LIQUIDITY
    const total = (await pair.read.getTokenTotalSupply([])) as bigint;
    console.log("Total supply:", total.toString());

    // Fetch and log pair balances before burn
    const pairBalanceToken0 = (await token0.read.getTokenBalanceOf([
      pairAddress,
    ])) as bigint;
    const pairBalanceToken1 = (await token1.read.getTokenBalanceOf([
      pairAddress,
    ])) as bigint;
    console.log(
      "Pair Balance token0 before burn:",
      pairBalanceToken0.toString(),
    );
    console.log(
      "Pair Balance token1 before burn:",
      pairBalanceToken1.toString(),
    );

    // Fetch and log user balances before burn
    let userBalanceToken0 = (await token0.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    let userBalanceToken1 = (await token1.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log(
      "User Balance token0 before burn:",
      userBalanceToken0.toString(),
    );
    console.log(
      "User Balance token1 before burn:",
      userBalanceToken1.toString(),
    );

    const userLpBalance = (await pair.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log(
      "Total LP balance for user smart account:",
      userLpBalance.toString(),
    );
    // Execute burn
    console.log("Executing burn...");
    // Send LP tokens to the user smart account
    const hash3 = await router.write.removeLiquidity(
      [pairAddress, smartAccount.address],
      {
        tokens: [
          {
            id: pairAddress,
            amount: BigInt(userLpBalance),
          },
        ],
      },
    );

    await waitTillCompleted(client, hash3);

    console.log("Burn executed.");

    // Log balances after burn
    const balanceToken0 = (await token0.read.getTokenBalanceOf([
      pairAddress,
    ])) as bigint;
    const balanceToken1 = (await token1.read.getTokenBalanceOf([
      pairAddress,
    ])) as bigint;
    console.log(
      "REMOVELIQUIDITY RESULT: Pair Balance token0 after burn:",
      balanceToken0.toString(),
    );
    console.log(
      "REMOVELIQUIDITY RESULT: Pair Balance token1 after burn:",
      balanceToken1.toString(),
    );

    userBalanceToken0 = (await token0.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    userBalanceToken1 = (await token1.read.getTokenBalanceOf([
      smartAccount.address,
    ])) as bigint;
    console.log(
      "REMOVELIQUIDITY RESULT: User Balance token0 after burn:",
      userBalanceToken0.toString(),
    );
    console.log(
      "REMOVELIQUIDITY RESULT: User Balance token1 after burn:",
      userBalanceToken1.toString(),
    );

    // Fetch and log reserves after burn
    const reserves = await pair.read.getReserves([]);
    console.log(
      "REMOVELIQUIDITY RESULT: Reserves from pair after burn:",
      // @ts-ignore
      reserves[0].toString(),
      // @ts-ignore
      reserves[1].toString(),
    );
  },
);
