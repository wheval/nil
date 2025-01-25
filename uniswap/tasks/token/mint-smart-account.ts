import { getContract } from "@nilfoundation/niljs";
import type { Address } from "abitype";
import { task } from "hardhat/config";
import { createSmartAccount } from "../basic/basic";
import { mintAndSendToken } from "../util/token-utils";

task(
  "mint-smart-account",
  "Mint token from two contracts and send it to a specified smart account",
)
  .addParam("token", "The contract address of the first token")
  .addParam("amount", "The amount of token to mint and send")
  .setAction(async (taskArgs, _) => {
    const smartAccountAddress = process.env.SMART_ACCOUNT_ADDR as
      | Address
      | undefined;

    if (!smartAccountAddress) {
      throw new Error("SMART_ACCOUNT_ADDR is not set in environment variables");
    }

    const smartAccount = await createSmartAccount();

    // Destructure parameters for clarity
    const mintAmount = BigInt(taskArgs.amount);
    const tokenAddress = taskArgs.token;

    console.log(
      `Starting mint and transfer process for tokens ${tokenAddress}`,
    );
    // Mint and send token for both contracts using the refactored utility function
    await mintAndSendToken({
      smartAccount: smartAccount,
      contractAddress: tokenAddress,
      smartAccountAddress,
      mintAmount,
    });

    const TokenJson = require("../../artifacts/contracts/Token.sol/Token.json");
    const contract = getContract({
      abi: TokenJson.abi,
      address: tokenAddress,
      client: smartAccount.client,
      smartAccount: smartAccount,
      externalInterface: {
        signer: smartAccount.signer,
        methods: [],
      },
    });

    // Verify recipient balances
    const balance = await contract.read.getTokenBalanceOf([
      smartAccount.address,
    ]);

    console.log(`Recipient balance after transfer: ${balance}`);
  });
