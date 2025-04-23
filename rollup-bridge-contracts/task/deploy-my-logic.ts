import type { Abi } from "abitype";
import { task } from "hardhat/config";
import {
    FaucetClient,
    HttpTransport,
    LocalECDSAKeySigner,
    PublicClient,
    SmartAccountV1,
    convertEthToWei,
    Transaction,
    generateRandomPrivateKey,
    waitTillCompleted,
} from "@nilfoundation/niljs";
import { loadNilSmartAccount } from "./nil-smart-account";
import { L2NetworkConfig, loadNilNetworkConfig, saveNilNetworkConfig } from "../deploy/config/config-helper";
import { decodeFunctionResult, encodeFunctionData } from "viem";

// npx hardhat deploy-my-logic --networkname local
task("deploy-my-logic", "Deploys MyLogic contract on Nil Chain")
    .addParam("networkname", "The network to use") // Mandatory parameter
    .setAction(async (taskArgs) => {

        // Dynamically load artifacts
        const MyLogicJson = await import("../artifacts/contracts/bridge/l2/MyLogic.sol/MyLogic.json");
        const TransparentUpgradeableProxy = await import("../artifacts/contracts/common/TransparentUpgradeableProxy.sol/MyTransparentUpgradeableProxy.json");
        const ProxyAdmin = await import("../artifacts/node_modules/@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol/ProxyAdmin.json");

        if (!MyLogicJson || !MyLogicJson.default || !MyLogicJson.default.abi || !MyLogicJson.default.bytecode) {
            throw Error(`Invalid L2ETHBridge ABI`);
        }

        const networkName = taskArgs.networkname;
        console.log(`Running task on network: ${networkName}`);

        const deployerAccount = await loadNilSmartAccount();

        if (!deployerAccount) {
            throw Error(`Invalid Deployer SmartAccount`);
        }

        const balance = await deployerAccount.getBalance();

        console.log(`smart-contract${deployerAccount.address} is on shard: ${deployerAccount.shardId} with balance: ${balance}`);

        if (!(balance > BigInt(0))) {
            throw Error(`Insufficient or Zero balance for smart-account: ${deployerAccount.address}`);
        }

        // save the nilMessageTree Address in the json config for l2
        const l2NetworkConfig: L2NetworkConfig = loadNilNetworkConfig(networkName);

        const { address: l2EthBridgeImplementationAddress, hash: l2EthBridgeImplementationDeploymentTxHash } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: MyLogicJson.default.bytecode,
            abi: MyLogicJson.default.abi,
            args: [],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: BigInt("19340180000000"),
        });

        console.log(`address from deployment is: ${l2EthBridgeImplementationAddress}`);
        await waitTillCompleted(deployerAccount.client, l2EthBridgeImplementationDeploymentTxHash);
        console.log("✅ Logic Contract deployed at:", l2EthBridgeImplementationDeploymentTxHash);

        if (!l2EthBridgeImplementationDeploymentTxHash) {
            throw Error(`Invalid transaction output from deployContract call for L2ETHBridge Contract`);
        }

        if (!l2EthBridgeImplementationAddress) {
            throw Error(`Invalid address output from deployContract call for L2ETHBridge Contract`);
        }

        console.log(`NilMessageTree contract deployed at address: ${l2EthBridgeImplementationAddress} and with transactionHash: ${l2EthBridgeImplementationDeploymentTxHash}`);

        l2NetworkConfig.l2ETHBridgeConfig.l2ETHBridgeContracts.l2ETHBridgeImplementation = l2EthBridgeImplementationAddress;

        const initData = encodeFunctionData({
            abi: MyLogicJson.default.abi,
            functionName: "initialize",
            args: [999],
        });

        const { address: addressProxy, hash: hashProxy } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: TransparentUpgradeableProxy.default.bytecode,
            abi: TransparentUpgradeableProxy.default.abi,
            args: [l2EthBridgeImplementationAddress, deployerAccount.address, initData],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: convertEthToWei(0.001),
        });
        await waitTillCompleted(deployerAccount.client, hashProxy);
        console.log("✅ Transparent Proxy Contract deployed at:", addressProxy);

        l2NetworkConfig.l2ETHBridgeConfig.l2ETHBridgeContracts.l2ETHBridgeProxy = addressProxy;

        console.log("Waiting 5 seconds...");
        await new Promise((res) => setTimeout(res, 5000));

    });
