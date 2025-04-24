import { task } from "hardhat/config";

task("deploy-factory").setAction(async (taskArgs, hre) => {
  const smartAccount = await hre.nil.getSmartAccount();

  const contract = await hre.nil.deployContract(
    "UniswapV2Factory",
    [smartAccount.address],
    {},
  );
  console.log(
    "Uniswap factory contract deployed at address: " + contract.address,
  );
});
