import { expect } from "chai";
import hre from "hardhat";
import "@nilfoundation/hardhat-nil-plugin";
import { waitTillCompleted } from "@nilfoundation/niljs";

describe("Guard check test", () => {
  it("positive_scenario", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const guardCheckChild = await hre.nil.deployContract("GuardCheckChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const guardCheckParent = await hre.nil.deployContract("GuardCheckParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await guardCheckParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await guardCheckParent.write.guardCheck([guardCheckChild.address, 1000, 4000]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(4000n);
    expect(await guardCheckChild.read.executed([])).to.equal(true);
  });

  it("require_failed_insufficient_balance_scenario", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const guardCheckChild = await hre.nil.deployContract("GuardCheckChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const guardCheckParent = await hre.nil.deployContract("GuardCheckParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await guardCheckParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await guardCheckParent.write.guardCheck([guardCheckChild.address, 6000, 0]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);
    expect(await guardCheckChild.read.executed([])).to.equal(false);
  });

  it("revert_failed_incorrect_amount_scenario", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const guardCheckChild = await hre.nil.deployContract("GuardCheckChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const guardCheckParent = await hre.nil.deployContract("GuardCheckParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await guardCheckParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await guardCheckParent.write.guardCheck([guardCheckChild.address, 500, 0]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);
    expect(await guardCheckChild.read.executed([])).to.equal(false);
  });

  it("assert_failed_child_contract_not_executed", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const guardCheckChild = await hre.nil.deployContract("GuardCheckChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const guardCheckParent = await hre.nil.deployContract("GuardCheckParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await guardCheckParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await guardCheckParent.write.guardCheck([guardCheckChild.address, 2000, 0]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await guardCheckParent.read.exampleBalance([])).to.equal(5000n);
    expect(await guardCheckChild.read.executed([])).to.equal(false);
  });
});
