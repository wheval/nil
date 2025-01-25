import { expect } from "chai";
import "@nomicfoundation/hardhat-ethers";
import { deployNilContract } from "../src/deploy";
import { createSmartAccount } from "../src/smart-account";
import type { Abi } from "abitype";
import { waitTillCompleted } from "@nilfoundation/niljs";

describe("Token contract", () => {
  it("Should deploy, transfer token, and verify balances", async () => {
    // Initialize values for testing
    const initialSupply = 10000;
    const smartAccount = await createSmartAccount({faucetDeposit: true});

    // Deploy the Token contract with initial supply and name
    const TokenJson = require("../artifacts/contracts/Token.sol/Token.json");
    const {contract: tokenBase, address: tokenBaseAddr} =
      await deployNilContract(
        smartAccount,
        TokenJson.abi as Abi,
        TokenJson.bytecode,
        [initialSupply],
        smartAccount.shardId,
      );

    console.log("Token deployed at:", tokenBaseAddr);

    const balance = await tokenBase.read.getOwnTokenBalance([]);
    console.log("Token balance:", balance.toString());
    expect(balance).to.equal(initialSupply);

    // Fetch and verify the token ID
    const tokenId = await tokenBase.read.getTokenId([]);
    console.log("Token ID:", tokenId);

    // Compare the expected and actual token IDs
    expect(tokenId.toString().toLowerCase()).to.equal(tokenBaseAddr.toString());

    // Deploy the IncrementerPayable contract
    const IncrementerJson = require("../artifacts/contracts/IncrementerPayable.sol/IncrementerPayable.json");
    const {contract: incrementer, address: incrementerAddr} =
      await deployNilContract(
        smartAccount,
        IncrementerJson.abi as Abi,
        IncrementerJson.bytecode,
        [],
        smartAccount.shardId,
      );
    console.log("IncrementerPayable deployed at:", incrementerAddr);

    // Transfer token from the Token contract to the IncrementerPayable contract
    const transferAmount = 500;
    const tx = await tokenBase.write.transferToken([incrementerAddr, tokenId, transferAmount]);
    await waitTillCompleted(smartAccount.client, tx);

    // After the transfer, verify that the balance in the Token contract has decreased
    const updatedBalance = await tokenBase.read.getOwnTokenBalance([]);
    console.log("Updated Token balance:", updatedBalance.toString());
    expect(updatedBalance).to.equal(initialSupply - transferAmount);

    // Fetch and verify that the total supply remains unchanged
    const totalSupply = await tokenBase.read.getTokenTotalSupply([]);
    console.log("Total Supply:", totalSupply.toString());
    expect(totalSupply).to.equal(initialSupply);

    // Verify the balance of the IncrementerPayable contract using getTokenBalanceOf
    const incrementerBalance = await tokenBase.read.getTokenBalanceOf([incrementerAddr]);
    console.log("IncrementerPayable balance:", incrementerBalance.toString());
    expect(incrementerBalance).to.equal(transferAmount);
  });
});
