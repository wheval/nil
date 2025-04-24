import { expect } from "chai";
import "@nomicfoundation/hardhat-ethers";
import hre from "hardhat";
import { waitTillCompleted } from "@nilfoundation/niljs";
import "@nilfoundation/hardhat-nil-plugin";


describe("Caller and Incrementer contract interaction", () => {
  let callerAddress: string;
  let incrementerAddress: string;

  it("Should deploy Caller with shardId 2, deploy Incrementer with shardId 1, and call incrementer from caller", async () => {

    const smartAccount = await hre.nil.createSmartAccount({topUp: true});
    const client = await hre.nil.getPublicClient();

    // Deploy Caller contract
    const caller = await hre.nil.deployContract("Caller", [], {
      smartAccount: smartAccount,
      shardId: 2,
      feeCredit: 100_000_000_000_000n,
    })

    callerAddress = caller.address;
    console.log("Caller deployed at:", callerAddress);

    // Deploy Incrementer contract
    const incrementer = await hre.nil.deployContract("Incrementer", [], {
      shardId: 1,
      smartAccount: smartAccount,
    })
    incrementerAddress = incrementer.address;
    console.log("Incrementer deployed at:", incrementerAddress);

    // Call the Caller contract's call method with the Incrementer address
    const tx = await caller.write.call([incrementerAddress]);
    await waitTillCompleted(client, tx, {waitTillMainShard: true});

    // Check the value of Incrementer
    expect(await incrementer.read.getValue([])).to.equal(1n);
  });
});
