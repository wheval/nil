import { expect } from "chai";
import { deployNilContract } from "../src/deploy";
import { createSmartAccount } from "../src/smart-account";
import type { Abi } from "abitype";
import { waitTillCompleted } from "@nilfoundation/niljs";

describe("Requester and Increment contract interaction", () => {

  it("Should Deploy Requester with shardId 2, deploy Incrementer with shardId 1, and call incrementer from caller using sendRequest", async function () {
    this.timeout(1200000);
    // Set shardId to 2
    const smartAccount = await createSmartAccount({faucetDeposit: true});

    const SendRequestJson = require("../artifacts/contracts/SendRequest.sol/SendRequest.json");
    const {contract: sendRequester, address: callerAddr} =
      await deployNilContract(
        smartAccount,
        SendRequestJson.abi as Abi,
        SendRequestJson.bytecode,
        [],
        2,
      );

    // Set shardId back to 1
    const IncrementerJson = require("../artifacts/contracts/Incrementer.sol/Incrementer.json");
    const {contract: incrementer, address: incrementerAddr} =
      await deployNilContract(
        smartAccount,
        IncrementerJson.abi as Abi,
        IncrementerJson.bytecode,
        [],
        smartAccount.shardId,
      );

    // Deploy Incrementer contract
    console.log("Incrementer deployed at:", incrementerAddr);

    //Generate a random string
    const randomString = Math.random().toString(36).substring(7);
    // Increment the value
    const tx = await incrementer.write.increment([]);
    await waitTillCompleted(smartAccount.client, tx);
    // Check the value of Incrementer
    expect(await incrementer.read.getValue()).to.equal(1);

    // Call the Caller contract's call method with the Incrementer address
    const tx2 = await sendRequester.write.call([incrementerAddr, randomString]);
    await waitTillCompleted(smartAccount.client, tx2);

    const value = await sendRequester.read.values([randomString]);
    console.log("returned value: ", value);
    // New value should be 1
    expect(value).to.equal(1);
  });
});
