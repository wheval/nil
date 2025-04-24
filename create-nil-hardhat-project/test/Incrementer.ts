import { expect } from "chai";
import hre from "hardhat";
import { waitTillCompleted } from "@nilfoundation/niljs";
import "@nilfoundation/hardhat-nil-plugin";

describe("Incrementer contract", () => {
  it("Should increment the value", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const incrementer = await hre.nil.deployContract("Incrementer", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
      feeCredit: 100_000_000_000_000n,
    });

    // Initial value should be 0
    expect(await incrementer.read.getValue([])).to.equal(0n);

    // Increment the value
    const tx = await incrementer.write.increment([]);
    await waitTillCompleted(client, tx, { waitTillMainShard: true });

    // New value should be 1
    expect(await incrementer.read.getValue([])).to.equal(1n);

    // Increment the value again
    const tx2 = await incrementer.write.increment([]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    // New value should be 2
    expect(await incrementer.read.getValue([])).to.equal(2n);
  });
});
