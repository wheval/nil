import { expect } from "chai";
import hre from "hardhat";
import { waitTillCompleted } from "@nilfoundation/niljs";
import "@nilfoundation/hardhat-nil-plugin";

describe("Requester and Increment contract interaction", () => {

  it("Should Deploy Requester with shardId 2, deploy Incrementer with shardId 1, and call incrementer from caller using sendRequest", async function () {
    this.timeout(1200000);

    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const sendRequester = await hre.nil.deployContract("SendRequest", [], {
      smartAccount: smartAccount,
      shardId: 2,
      feeCredit: 100_000_000_000_000n,
    });

    console.log("SendRequest deployed at:", sendRequester.address);

    // Deploy Incrementer contract with shardId 1
    const incrementer = await hre.nil.deployContract("Incrementer", [], {
      smartAccount: smartAccount,
      shardId: 1,
    });

    console.log("Incrementer deployed at:", incrementer.address);

    //Generate a random string
    const randomString = Math.random().toString(36).substring(7);

    // Increment the value
    const tx = await incrementer.write.increment([]);
    await waitTillCompleted(client, tx, { waitTillMainShard: true });

    // Check the value of Incrementer
    expect(await incrementer.read.getValue([])).to.equal(1n);

    // Call the SendRequest contract's call method with the Incrementer address
    const tx2 = await sendRequester.write.call([incrementer.address, randomString]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    const value = await sendRequester.read.values([randomString]);
    console.log("returned value: ", value);

    // New value should be 1
    expect(value).to.equal(1n);
  });
});
