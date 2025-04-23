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

// npx hardhat deploy-l2-bridge-messenger --networkname local
task("deploy-l2-bridge-messenger", "Deploys L2BridgeMessenger contract on Nil Chain")
    .addParam("networkname", "The network to use") // Mandatory parameter
    .setAction(async (taskArgs) => {

        // Dynamically load artifacts
        const L2BridgeMessengerJson = await import("../artifacts/contracts/bridge/l2/L2BridgeMessenger.sol/L2BridgeMessenger.json");
        const TransparentUpgradeableProxy = await import("../artifacts/contracts/common/TransparentUpgradeableProxy.sol/MyTransparentUpgradeableProxy.json");
        const ProxyAdmin = await import("../artifacts/node_modules/@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol/ProxyAdmin.json");

        if (!L2BridgeMessengerJson || !L2BridgeMessengerJson.default || !L2BridgeMessengerJson.default.abi || !L2BridgeMessengerJson.default.bytecode) {
            throw Error(`Invalid L2BridgeMessengerJson ABI`);
        }

        const networkName = taskArgs.networkname;
        console.log(`Running task on network: ${networkName}`);

        const deployerAccount = await loadNilSmartAccount();

        if (!deployerAccount) {
            throw Error(`Invalid Deployer SmartAccount`);
        }

        const balance = await deployerAccount.getBalance();

        console.log(`smart-contract: ${deployerAccount.address} is on shard: ${deployerAccount.shardId} with balance: ${balance}`);

        if (!(balance > BigInt(0))) {
            throw Error(`Insufficient or Zero balance for smart-account: ${deployerAccount.address}`);
        }

        // save the L2BridgeMessenger Address in the json config for l2
        const l2NetworkConfig: L2NetworkConfig = loadNilNetworkConfig(networkName);

        const { address: nilMessengerImplementationAddress, hash: nilMessengerImplementationDeploymentTxHash } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: L2BridgeMessengerJson.default.bytecode,
            abi: L2BridgeMessengerJson.default.abi,
            args: [],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: BigInt("19340180000000"),
        });

        await waitTillCompleted(deployerAccount.client, nilMessengerImplementationDeploymentTxHash);
        console.log(`L2BridgeMessenger contractis deployed at: ${nilMessengerImplementationAddress} with transactionHash: ${nilMessengerImplementationDeploymentTxHash}`);

        if (!nilMessengerImplementationDeploymentTxHash) {
            throw Error(`Invalid transaction output from deployContract call for L2BridgeMessenger Contract`);
        }

        if (!nilMessengerImplementationAddress) {
            throw Error(`Invalid address output from deployContract call for L2BridgeMessenger Contract`);
        }

        console.log(`L2BridgeMessenger contract deployed at address: ${nilMessengerImplementationAddress} and with transactionHash: ${nilMessengerImplementationDeploymentTxHash}`);

        l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerContracts.l2BridgeMessengerImplementation = nilMessengerImplementationAddress;

        const initData = encodeFunctionData({
            abi: L2BridgeMessengerJson.default.abi,
            functionName: "initialize",
            args: [l2NetworkConfig.l2CommonConfig.owner, l2NetworkConfig.l2CommonConfig.admin,
            l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerDeployerConfig.relayerAddress,
            l2NetworkConfig.nilMessageTreeConfig.nilMessageTreeContracts.nilMessageTreeImplementationAddress,
            l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerDeployerConfig.messageExpiryDeltaValue],
        });

        const { address: addressProxy, hash: hashProxy } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: TransparentUpgradeableProxy.default.bytecode,
            abi: TransparentUpgradeableProxy.default.abi,
            args: [nilMessengerImplementationAddress, deployerAccount.address, initData],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: convertEthToWei(0.001),
        });
        await waitTillCompleted(deployerAccount.client, hashProxy);
        console.log("✅ Transparent Proxy Contract deployed at:", addressProxy);

        l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerContracts.l2BridgeMessengerProxy = addressProxy;

        console.log("Waiting 5 seconds...");
        await new Promise((res) => setTimeout(res, 5000));


        const getNilMessageTreeCallData = encodeFunctionData({
            abi: L2BridgeMessengerJson.default.abi,
            functionName: "nilMessageTree",
            args: [],
        });
        const getNilMessageTreeCallResult = await deployerAccount.client.call({
            to: addressProxy,
            from: deployerAccount.address,
            data: getNilMessageTreeCallData,
        }, "latest");

        console.log(`getNilMessageTreeCallResult is: ${JSON.stringify(getNilMessageTreeCallResult)}`);

        const result = decodeFunctionResult({
            abi: L2BridgeMessengerJson.default.abi,
            functionName: "nilMessageTree",
            data: getNilMessageTreeCallResult.data,
        });
        console.log("✅ NilMessageTree value in L2BridgeMessenger contract:", result);


        const getImplementationCallData = encodeFunctionData({
            abi: L2BridgeMessengerJson.default.abi,
            functionName: "getImplementation",
            args: [],
        });

        const getImplementationResult = await deployerAccount.client.call({
            to: addressProxy,
            data: getImplementationCallData,
            from: deployerAccount.address,
        }, "latest");

        console.log(`L2BridgeMessengerProxy has get-implementation-result: ${JSON.stringify(getImplementationResult)}`);

        const proxyImplementationAddress = decodeFunctionResult({
            abi: L2BridgeMessengerJson.default.abi,
            functionName: "getImplementation",
            data: getImplementationResult.data,
        }) as string;

        console.log("✅ proxyImplementationAddress Address:", proxyImplementationAddress);

        const fetchAdminCall = encodeFunctionData({
            abi: TransparentUpgradeableProxy.default.abi,
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
            abi: TransparentUpgradeableProxy.default.abi,
            functionName: "fetchAdmin",
            data: adminResult.data,
        }) as string;

        console.log("✅ ProxyAdmin Address:", proxyAdminAddress);

        l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerContracts.proxyAdmin = proxyAdminAddress;

        // Save the updated config
        saveNilNetworkConfig(networkName, l2NetworkConfig);
    });
