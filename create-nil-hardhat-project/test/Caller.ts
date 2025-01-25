import { expect } from "chai";
import "@nomicfoundation/hardhat-ethers";
import { deployNilContract } from "../src/deploy";
import type { Abi } from "abitype";
import { createSmartAccount } from "../src/smart-account";
import { waitTillCompleted } from "@nilfoundation/niljs";


describe("Caller and Incrementer contract interaction", () => {
  let callerAddress: string;
  let incrementerAddress: string;

  it("Should deploy Caller with shardId 2, deploy Incrementer with shardId 1, and call incrementer from caller", async () => {

    // Deploy Caller contract
    const smartAccount = await createSmartAccount({faucetDeposit: true});
    const AwaitJson = require("../artifacts/contracts/Caller.sol/Caller.json");

    const {contract: caller, address: callerAddr} =
      await deployNilContract(
        smartAccount,
        AwaitJson.abi as Abi,
        AwaitJson.bytecode,
        [],
        2,
        [],
      );

    callerAddress = callerAddr;
    console.log("Caller deployed at:", callerAddress);

    // Deploy Incrementer contract
    const IncrementerJson = require("../artifacts/contracts/Incrementer.sol/Incrementer.json");
    const {contract: incrementer, address: incrementerAddr} =
      await deployNilContract(
        smartAccount,
        IncrementerJson.abi as Abi,
        IncrementerJson.bytecode,
        [],
        1,
        [],
      );
    incrementerAddress = incrementerAddr;
    console.log("Incrementer deployed at:", incrementerAddress);

    // Call the Caller contract's call method with the Incrementer address
    const tx = await caller.write.call([incrementerAddress]);
    await waitTillCompleted(smartAccount.client, tx);

    // Check the value of Incrementer
    expect(await incrementer.read.getValue([])).to.equal(1);
  });
});
