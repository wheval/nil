import { task } from "hardhat/config";
import { createSmartAccount } from "./basic";

task("create-smart-account").setAction(async (taskArgs, _) => {
  const smartAccount = await createSmartAccount({
    faucetDeposit: true,
  });
  console.log("Smart account created: " + smartAccount.address);
});
