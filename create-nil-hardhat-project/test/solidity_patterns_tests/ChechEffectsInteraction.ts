import { expect } from "chai";
import hre from "hardhat";
import "@nomicfoundation/hardhat-ethers";
import "@nilfoundation/hardhat-nil-plugin";
import { waitTillCompleted } from "@nilfoundation/niljs";

describe("Check Effects Interaction test", () => {
  it("positive_scenario", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const ceiChild = await hre.nil.deployContract("CheckEffectsInteractionChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const ceiParent = await hre.nil.deployContract("CheckEffectsInteractionParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await ceiParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await ceiParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await ceiParent.write.checkEffectsInteraction([ceiChild.address, 500, true]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await ceiParent.read.exampleBalance([])).to.equal(4500n);
  });

  it("balance_not_changed_with_error_in_child_contract", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const ceiChild = await hre.nil.deployContract("CheckEffectsInteractionChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const ceiParent = await hre.nil.deployContract("CheckEffectsInteractionParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await ceiParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await ceiParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await ceiParent.write.checkEffectsInteraction([ceiChild.address, 500, false]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await ceiParent.read.exampleBalance([])).to.equal(5000n);
  });

  it("check_protects_contract_with_insufficient_balance", async () => {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const ceiChild = await hre.nil.deployContract("CheckEffectsInteractionChild", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const ceiParent = await hre.nil.deployContract("CheckEffectsInteractionParent", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx1 = await ceiParent.write.topUpBalance([5000]);
    await waitTillCompleted(client, tx1, { waitTillMainShard: true });

    expect(await ceiParent.read.exampleBalance([])).to.equal(5000n);

    const tx2 = await ceiParent.write.checkEffectsInteraction([ceiChild.address, 5500, true]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await ceiParent.read.exampleBalance([])).to.equal(5000n);
  });
});
