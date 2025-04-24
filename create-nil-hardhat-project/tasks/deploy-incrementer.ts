import { task } from "hardhat/config";

task("deploy-incrementer").setAction(async (taskArgs, hre) => {
  const contract = await hre.nil.deployContract("Incrementer", []);

  console.log("Incrementer contract deployed at address: " + contract.address);

  await contract.write.increment([]);

  const value = await contract.read.getValue([]);

  console.log("Incrementer contract value: " + value);
});
