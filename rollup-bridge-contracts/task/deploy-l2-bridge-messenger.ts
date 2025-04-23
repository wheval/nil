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
import * as L2BridgeMessengerJson from "../artifacts/contracts/bridge/l2/L2BridgeMessenger.sol/L2BridgeMessenger.json";
import * as TransparentUpgradeableProxy from "../artifacts/contracts/common/TransparentUpgradeableProxy.sol/MyTransparentUpgradeableProxy.json";
import * as ProxyAdmin from "../artifacts/node_modules/@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol/ProxyAdmin.json";
import { decodeFunctionResult, encodeFunctionData } from "viem";

// npx hardhat deploy-l2-bridge-messenger --networkname local
task("deploy-l2-bridge-messenger", "Deploys L2BridgeMessenger contract on Nil Chain")
    .addParam("networkname", "The network to use") // Mandatory parameter
    .setAction(async (taskArgs) => {

        if (!L2BridgeMessengerJson || !L2BridgeMessengerJson.abi || !L2BridgeMessengerJson.bytecode) {
            throw Error(`Invalid L2BridgeMessengerJson ABI`);
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

        const { address: nilMessengerImplementationAddress, hash: nilMessengerImplementationDeploymentTxHash } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: L2BridgeMessengerJson.bytecode,
            abi: L2BridgeMessengerJson.abi,
            args: [],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: BigInt("19340180000000"),
        });

        console.log(`address from deployment is: ${nilMessengerImplementationAddress}`);
        await waitTillCompleted(deployerAccount.client, nilMessengerImplementationDeploymentTxHash);
        console.log("✅ Logic Contract deployed at:", nilMessengerImplementationDeploymentTxHash);

        if (!nilMessengerImplementationDeploymentTxHash) {
            throw Error(`Invalid transaction output from deployContract call for NilMessageTree Contract`);
        }

        if (!nilMessengerImplementationAddress) {
            throw Error(`Invalid address output from deployContract call for NilMessageTree Contract`);
        }

        console.log(`NilMessageTree contract deployed at address: ${nilMessengerImplementationAddress} and with transactionHash: ${nilMessengerImplementationDeploymentTxHash}`);

        l2NetworkConfig.l2BridgeMessenger.l2BridgeMessengerContracts.l2BridgeMessengerImplementation = nilMessengerImplementationAddress;

        const initData = encodeFunctionData({
            abi: L2BridgeMessengerJson.abi,
            functionName: "initialize",
            args: [l2NetworkConfig.l2Common.owner, l2NetworkConfig.l2Common.admin,
            l2NetworkConfig.l2BridgeMessenger.l2BridgeMessengerDeployerConfig.relayerAddress,
            l2NetworkConfig.nilMessageTree.nilMessageTreeContracts.nilMessageTreeImplementationAddress,
            l2NetworkConfig.l2BridgeMessenger.l2BridgeMessengerDeployerConfig.messageExpiryDeltaValue],
        });

        console.log("Deploying L2BridgeMessenger with args:");
        console.log("implementationAddress:", nilMessengerImplementationAddress);
        console.log("owner:", deployerAccount.address);
        console.log("L2BridgeMessenger InitData:", initData);

        const { address: addressProxy, hash: hashProxy } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: TransparentUpgradeableProxy.bytecode,
            abi: TransparentUpgradeableProxy.abi,
            args: [nilMessengerImplementationAddress, deployerAccount.address, initData],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: convertEthToWei(0.001),
        });
        await waitTillCompleted(deployerAccount.client, hashProxy);
        console.log("✅ Transparent Proxy Contract deployed at:", addressProxy);

        l2NetworkConfig.l2BridgeMessenger.l2BridgeMessengerContracts.l2BridgeMessengerProxy = addressProxy;

        console.log("Waiting 5 seconds...");
        await new Promise((res) => setTimeout(res, 5000));

        const fetchImplementationCall = encodeFunctionData({
            abi: TransparentUpgradeableProxy.abi,
            functionName: "fetchImplementation",
            args: [],
        });

        const fetchImplementationResult = await deployerAccount.client.call({
            to: addressProxy,
            data: fetchImplementationCall,
            from: deployerAccount.address,
        }, "latest");

        console.log(`L2BridgeMessengerProxy has fetch-implementation-result: ${JSON.stringify(fetchImplementationResult)}`);

        const proxyImplementationAddress = decodeFunctionResult({
            abi: TransparentUpgradeableProxy.abi,
            functionName: "fetchImplementation",
            data: fetchImplementationResult.data,
        }) as string;

        console.log("✅ proxyImplementationAddress Address:", proxyImplementationAddress);

        const fetchAdminCall = encodeFunctionData({
            abi: TransparentUpgradeableProxy.abi,
            functionName: "fetchAdmin",
            args: [],
        });

        const adminResult = await deployerAccount.client.call({
            to: addressProxy,
            data: fetchAdminCall,
            from: deployerAccount.address,
        }, "latest");

        console.log(`L2BridgeMessengerProxy has admin-result: ${JSON.stringify(adminResult)}`);

        const proxyAdminAddress = decodeFunctionResult({
            abi: TransparentUpgradeableProxy.abi,
            functionName: "fetchAdmin",
            data: adminResult.data,
        }) as string;

        console.log("✅ ProxyAdmin Address:", proxyAdminAddress);

        l2NetworkConfig.l2BridgeMessenger.l2BridgeMessengerContracts.proxyAdmin = proxyAdminAddress;

        // Save the updated config
        saveNilNetworkConfig(networkName, l2NetworkConfig);
    });
