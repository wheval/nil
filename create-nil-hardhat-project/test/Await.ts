import { expect } from "chai";
import hre from "hardhat";
import { waitTillCompleted } from "@nilfoundation/niljs";
import "@nilfoundation/hardhat-nil-plugin";

describe("Caller and Incrementer contract interaction", () => {
  let callerAddress: string;
  let incrementerAddress: string;

  it("Should deploy Caller with shardId 2, deploy Incrementer with shardId 1, and call incrementer from caller using await", async function () {
    this.timeout(120000);

    // Deploy Caller contract
    const smartAccount = await hre.nil.createSmartAccount({topUp: true});
    const client = await hre.nil.getPublicClient();

    const awaiter = await hre.nil.deployContract("Await", [], {
      shardId: 2,
      smartAccount: smartAccount,
      feeCredit: 100_000_000_000_000n,
    })

    callerAddress = awaiter.address;
    console.log("Caller deployed at:", callerAddress);

    // Deploy Incrementer contract
    const incrementer = await hre.nil.deployContract("Incrementer", [], {
      shardId: 1,
      smartAccount: smartAccount,
    })
    incrementerAddress = incrementer.address;
    console.log("Incrementer deployed at:", incrementerAddress);

    // Increment the value
    const tx = await incrementer.write.increment([]);
    await waitTillCompleted(client, tx, {waitTillMainShard: true})

    // Check the value of Incrementer
    expect(await incrementer.read.getValue([])).to.equal(1n);

    // Call the Caller contract's call method with the Incrementer address
    const tx2 = await awaiter.write.call([incrementerAddress]);
    await waitTillCompleted(client, tx2, {waitTillMainShard: true})

    let value = await awaiter.read.result([]);
    console.log("returned value: ", value)

    // New value should be 1
    expect(value).to.equal(1n);
  });
});

