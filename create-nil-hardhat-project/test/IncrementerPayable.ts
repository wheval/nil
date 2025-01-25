import { expect } from "chai";
import { deployNilContract } from "../src/deploy";
import type { Abi } from "abitype";
import { createSmartAccount } from "../src/smart-account";
import { waitTillCompleted } from "@nilfoundation/niljs";


describe("Incrementer contract", () => {
  it("Should increment the value", async () => {
    const smartAccount = await createSmartAccount({faucetDeposit: true});

    const IncrementerJson = require("../artifacts/contracts/IncrementerPayable.sol/IncrementerPayable.json");
    const {contract: incrementer, address: incrementerAddr} =
      await deployNilContract(
        smartAccount,
        IncrementerJson.abi as Abi,
        IncrementerJson.bytecode,
        [],
        smartAccount.shardId,
      );

    // Initial value should be 0
    expect(await incrementer.read.getValue([])).to.equal(0);

    // Increment the value
    const tx = await incrementer.write.increment([], {value: 1n});
    await waitTillCompleted(smartAccount.client, tx);

    // New value should be 1
    expect(await incrementer.read.getValue([])).to.equal(1);

    // Increment the value again
    const tx2 = await incrementer.write.increment([], {value: 1n});
    await waitTillCompleted(smartAccount.client, tx2);

    // New value should be 2
    expect(await incrementer.read.getValue([])).to.equal(2);
  });
});
