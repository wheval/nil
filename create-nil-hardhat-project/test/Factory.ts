import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat";
import { waitTillCompleted } from "@nilfoundation/niljs";
import "@nilfoundation/hardhat-nil-plugin";

describe("Factory contract interaction", function () {
  let incrementerAddress: `0x${string}`;

  it("Should deploy Incrementer using Factory, increment its value and verify", async function () {
    this.timeout(120000);

    // Deploy Factory contract
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const factory = await hre.nil.deployContract("Factory", [], {
      smartAccount: smartAccount,
      feeCredit: 100_000_000_000_000n,
    });
    console.log("Factory deployed at:", factory.address);

    // Retrieve Incrementer bytecode from Hardhat
    const Incrementer = await ethers.getContractFactory("Incrementer");
    const incrementerBytecode = Incrementer.bytecode;

    const shardId = 2; // Set shardId to 2
    const randomSalt = Math.floor(Math.random() * 10000) + 1;

    // Deploy Incrementer contract using Factory's deploy method
    const tx = await factory.write.deploy(["Incrementer", shardId, incrementerBytecode, randomSalt], {feeCredit: 100_000_000_000_000n});
    await waitTillCompleted(client, tx, { waitTillMainShard: true });

    // Get the deployed Incrementer contract address
    incrementerAddress = await factory.read.getContractAddress(["Incrementer"]) as `0x${string}`;
    console.log("Incrementer deployed at:", incrementerAddress);

    // Attach to the deployed Incrementer contract
    const incrementer = await hre.nil.getContractAt("Incrementer", incrementerAddress, {
      smartAccount: smartAccount,
    });

    // Call increment function
    const tx2 = await incrementer.write.increment([]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    // Check the value of Incrementer using getValue
    const value = await incrementer.read.getValue([]);
    console.log("Incrementer value:", value);

    // Expect the value to be 1 after increment
    expect(value).to.equal(1n);
  });
});
