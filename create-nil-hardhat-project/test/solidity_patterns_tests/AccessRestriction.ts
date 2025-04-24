import hre from "hardhat";
import "@nilfoundation/hardhat-nil-plugin";
import {expect} from "chai";

describe("Access Restriction test", () => {
  it("positive_scenario", async () => {

    const smartAccount = await hre.nil.createSmartAccount({topUp: true});

    const ar = await hre.nil.deployContract("AccessRestriction", [], {smartAccount: smartAccount});

    hre.nil.deployContract
    const owner = (await ar.read.owner([])) as `0x${string}`;
    expect(await ar.read.controlValue([])).to.equal(0)
    await ar.write.addAdmin([owner])
    await ar.write.accessRestrictionAction([])
    expect(await ar.read.controlValue([])).to.equal(1)
  });

  it("access_restricted_scenario", async () => {
    const ar = await hre.nil.deployContract("AccessRestriction", []);

    expect(await ar.read.controlValue([])).to.equal(0)
    await ar.write.accessRestrictionAction([])
    expect(await ar.read.controlValue([])).to.equal(0)
  });

  it("admin_excluded_from_pool", async function () {
    this.timeout(60000);

    const ar = await hre.nil.deployContract("AccessRestriction", []);
    const owner = (await ar.read.owner([])) as `0x${string}`;

    await ar.write.addAdmin([owner])
    await ar.write.accessRestrictionAction([])
    await ar.write.removeAdmin([owner])
    await ar.write.accessRestrictionAction([])
    expect(await ar.read.controlValue([])).to.equal(1)
  });
})
