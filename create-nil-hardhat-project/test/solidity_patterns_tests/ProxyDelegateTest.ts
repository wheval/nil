import { expect } from "chai";
import hre from "hardhat";
import "@nilfoundation/hardhat-nil-plugin";
import { waitTillCompleted } from "@nilfoundation/niljs";

describe("Proxy delegate pattern test", function () {
  let delegateAddress: `0x${string}`;
  let proxyAddress: `0x${string}`;
  let proxy: any;
  let delegate: any;

  before(async function () {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    delegate = await hre.nil.deployContract("Delegate", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });
    delegateAddress = delegate.address;

    proxy = await hre.nil.deployContract("Proxy", [delegateAddress], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });
    proxyAddress = proxy.address;
  });

  it("delegate positive scenario", async function () {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    // Get Proxy with Delegate ABI
    const proxyWithDelegateABI = await hre.nil.getContractAt("Delegate", proxyAddress, {
      smartAccount: smartAccount
    });

    const tx = await proxyWithDelegateABI.write.setValue([42]);
    await waitTillCompleted(client, tx, { waitTillMainShard: true });

    expect(await proxyWithDelegateABI.read.getValue([])).to.equal(42n);
  });

  it("change delegate scenario", async function () {
    const smartAccount = await hre.nil.createSmartAccount({ topUp: true });
    const client = await hre.nil.getPublicClient();

    const delegate2 = await hre.nil.deployContract("Delegate2", [], {
      smartAccount: smartAccount,
      shardId: smartAccount.shardId,
    });

    const tx = await proxy.write.upgradeTo([delegate2.address]);
    await waitTillCompleted(client, tx, { waitTillMainShard: true });

    // Get Proxy with Delegate2 ABI
    const proxyWithDelegate2ABI = await hre.nil.getContractAt("Delegate2", proxyAddress, {
      smartAccount: smartAccount
    });

    const tx2 = await proxyWithDelegate2ABI.write.setValue([42]);
    await waitTillCompleted(client, tx2, { waitTillMainShard: true });

    expect(await proxyWithDelegate2ABI.read.getValue([])).to.equal(142n);
  });
});
