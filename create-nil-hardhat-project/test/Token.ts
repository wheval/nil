import { expect } from "chai";
import "@nomicfoundation/hardhat-ethers";
import hre from "hardhat";
import { waitTillCompleted } from "@nilfoundation/niljs";
import "@nilfoundation/hardhat-nil-plugin";

describe("Token contract", () => {
  it("Should deploy, transfer token, and verify balances", async () => {
    // Initialize values for testing
    const initialSupply = 10000;
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    // Deploy the Token contract with initial supply and name
    const token = await hre.nil.deployContract("Token", [initialSupply], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
      feeCredit: 100_000_000_000_000n,
    });

    console.log("Token deployed at:", token.address);

    const balance = await token.read.getOwnTokenBalance([]) as bigint;
    console.log("Token balance:", balance.toString());
    expect(balance).to.equal(BigInt(initialSupply));

    // Fetch and verify the token ID
    const tokenId = await token.read.getTokenId([]) as string;
    console.log("Token ID:", tokenId);

    // Compare the expected and actual token IDs
    expect(tokenId.toString().toLowerCase()).to.equal(token.address.toString());

    // Deploy the IncrementerPayable contract
    const incrementer = await hre.nil.deployContract("IncrementerPayable", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });
    console.log("IncrementerPayable deployed at:", incrementer.address);

    // Transfer token from the Token contract to the IncrementerPayable contract
    const transferAmount = 500;
    const tx = await token.write.transferToken([incrementer.address, tokenId, transferAmount]);
    await waitTillCompleted(client, tx, { waitTillMainShard: true });

    // After the transfer, verify that the balance in the Token contract has decreased
    const updatedBalance = await token.read.getOwnTokenBalance([]) as bigint;
    console.log("Updated Token balance:", updatedBalance.toString());
    expect(updatedBalance).to.equal(BigInt(initialSupply - transferAmount));

    // Fetch and verify that the total supply remains unchanged
    const totalSupply = await token.read.getTokenTotalSupply([]) as bigint;
    console.log("Total Supply:", totalSupply.toString());
    expect(totalSupply).to.equal(BigInt(initialSupply));

    // Verify the balance of the IncrementerPayable contract using getTokenBalanceOf
    const incrementerBalance = await token.read.getTokenBalanceOf([incrementer.address]) as bigint;
    console.log("IncrementerPayable balance:", incrementerBalance.toString());
    expect(incrementerBalance).to.equal(BigInt(transferAmount));
  });
});
