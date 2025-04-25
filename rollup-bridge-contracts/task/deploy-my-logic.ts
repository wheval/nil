import type { Abi } from "abitype";
import { task } from "hardhat/config";
import {
    FaucetClient,
    HttpTransport,
    LocalECDSAKeySigner,
    PublicClient,
    SmartAccountV1,
    getContract,
    convertEthToWei,
    Transaction,
    generateRandomPrivateKey,
    waitTillCompleted,
} from "@nilfoundation/niljs";
import { loadNilSmartAccount } from "./nil-smart-account";
import { decodeFunctionResult, encodeFunctionData } from "viem";

// npx hardhat deploy-my-logic
task("deploy-my-logic", "Deploys MyLogic contract on Nil Chain")
    .setAction(async (taskArgs) => {

        // Dynamically load artifacts
        const MyLogicJson = await import("../artifacts/contracts/bridge/l2/MyLogic.sol/MyLogic.json");
        const TransparentUpgradeableProxy = await import("../artifacts/contracts/common/TransparentUpgradeableProxy.sol/MyTransparentUpgradeableProxy.json");
        const ProxyAdmin = await import("../artifacts/node_modules/@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol/ProxyAdmin.json");

        if (!MyLogicJson || !MyLogicJson.default || !MyLogicJson.default.abi || !MyLogicJson.default.bytecode) {
            throw Error(`Invalid myLogic ABI`);
        }

        const deployerAccount = await loadNilSmartAccount();

        if (!deployerAccount) {
            throw Error(`Invalid Deployer SmartAccount`);
        }

        const balance = await deployerAccount.getBalance();

        console.log(`smart-contract${deployerAccount.address} is on shard: ${deployerAccount.shardId} with balance: ${balance}`);

        if (!(balance > BigInt(0))) {
            throw Error(`Insufficient or Zero balance for smart-account: ${deployerAccount.address}`);
        }

        const { address: myLogicImplementationAddress, hash: myLogicImplementationDeploymentTxHash } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: MyLogicJson.default.bytecode,
            abi: MyLogicJson.default.abi,
            args: [],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: BigInt("19340180000000"),
        });

        console.log(`address from deployment is: ${myLogicImplementationAddress}`);
        await waitTillCompleted(deployerAccount.client, myLogicImplementationDeploymentTxHash);
        console.log("✅ Logic Contract deployed at:", myLogicImplementationDeploymentTxHash);

        if (!myLogicImplementationDeploymentTxHash) {
            throw Error(`Invalid transaction output from deployContract call for myLogic Contract`);
        }

        if (!myLogicImplementationAddress) {
            throw Error(`Invalid address output from deployContract call for myLogic Contract`);
        }

        console.log(`NilMessageTree contract deployed at address: ${myLogicImplementationAddress} and with transactionHash: ${myLogicImplementationDeploymentTxHash}`);


        const initData = encodeFunctionData({
            abi: MyLogicJson.default.abi,
            functionName: "initialize",
            args: [999],
        });

        const { address: addressProxy, hash: hashProxy } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: TransparentUpgradeableProxy.default.bytecode,
            abi: TransparentUpgradeableProxy.default.abi,
            args: [myLogicImplementationAddress, deployerAccount.address, initData],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: convertEthToWei(0.001),
        });
        await waitTillCompleted(deployerAccount.client, hashProxy);
        console.log("✅ Transparent Proxy Contract deployed at:", addressProxy);

        console.log("Waiting 5 seconds...");
        await new Promise((res) => setTimeout(res, 10000));

        console.log(`abi is: ${JSON.stringify(MyLogicJson.default.abi)}`);

        const myLogicContractInstance = getContract({
            abi: MyLogicJson.default.abi,
            address: addressProxy,
            client: deployerAccount.client,
            smartAccount: deployerAccount,
        });

        console.log("Properties of myLogicContractInstance:", Object.keys(myLogicContractInstance.read));

        const value = await myLogicContractInstance.read.getImplementation();

        console.log(`value from MyLogic is: ${value}`);
    });
