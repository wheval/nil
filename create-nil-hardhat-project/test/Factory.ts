import { expect } from "chai";
import { ethers } from "hardhat";
import { deployNilContract } from "../src/deploy";
import type { Abi } from "abitype";
import { createSmartAccount } from "../src/smart-account";
import { getContract, waitTillCompleted } from "@nilfoundation/niljs";

describe("Factory contract interaction", function () {
  let incrementerAddress: string;

  it("Should deploy Incrementer usinf Factory, increment its value and verify", async function () {
    this.timeout(120000);

    // Deploy Factory contract
    const smartAccount = await createSmartAccount({faucetDeposit: true});
    const FactoryJson = require("../artifacts/contracts/Factory.sol/Factory.json");

    const { contract: factory, address: callerAddr } = await deployNilContract(
      smartAccount,
      FactoryJson.abi as Abi,
      FactoryJson.bytecode,
      [],
      smartAccount.shardId,
      [],
    );
    console.log("Factory deployed at:", callerAddr);


    // Retrieve Incrementer bytecode from Hardhat
    const Incrementer = await ethers.getContractFactory("Incrementer");
    const incrementerBytecode = Incrementer.bytecode;

    const shardId = 2; // Set shardId to 2
    const randomSalt = Math.floor(Math.random() * 10000) + 1;

    // Deploy Incrementer contract using Factory's deploy method
    const tx = await factory.write.deploy(["Incrementer", shardId, incrementerBytecode, randomSalt]);
    await waitTillCompleted(smartAccount.client, tx);

    // Get the deployed Incrementer contract address
    incrementerAddress = await factory.read.getContractAddress(["Incrementer"]);
    console.log("Incrementer deployed at:", incrementerAddress);

    // Attach to the deployed Incrementer contract
    const IncrementerJson = require("../artifacts/contracts/Incrementer.sol/Incrementer.json");
    const incrementer = getContract({
      abi: IncrementerJson.abi,
      address: incrementerAddress,
      client: smartAccount.client,
      smartAccount: smartAccount,
      externalInterface: {
        signer: smartAccount.signer,
        methods: [],
      },
    });

    // Call increment function
    const tx2 = await incrementer.write.increment([]);
    await waitTillCompleted(smartAccount.client, tx2);

    // Check the value of Incrementer using getValue
    const value = await incrementer.read.getValue([]);
    console.log("Incrementer value:", value);

    // Expect the value to be 1 after increment
    expect(value).to.equal(1);
  });
});
