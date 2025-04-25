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

// npx hardhat deploy-l2-enshrined-token-bridge --networkname local
task("deploy-l2-enshrined-token-bridge", "Deploys L2EnshrinedTokenBridge contract on Nil Chain")
    .addParam("networkname", "The network to use") // Mandatory parameter
    .setAction(async (taskArgs) => {

        // Dynamically load artifacts
        const L2EnshrinedTokenBridgeJson = await import("../artifacts/contracts/bridge/l2/L2EnshrinedTokenBridge.sol/L2EnshrinedTokenBridge.json");
        const TransparentUpgradeableProxy = await import("../artifacts/contracts/common/TransparentUpgradeableProxy.sol/MyTransparentUpgradeableProxy.json");
        const ProxyAdmin = await import("../artifacts/node_modules/@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol/ProxyAdmin.json");

        if (!L2EnshrinedTokenBridgeJson || !L2EnshrinedTokenBridgeJson.default || !L2EnshrinedTokenBridgeJson.default.abi || !L2EnshrinedTokenBridgeJson.default.bytecode) {
            throw Error(`Invalid L2EnshrinedTokenBridge ABI`);
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

        const { address: l2EnshrinedTokenBridgeImplAddress, hash: l2EnshrinedTokenBridgeImplDepTxHash } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: L2EnshrinedTokenBridgeJson.default.bytecode,
            abi: L2EnshrinedTokenBridgeJson.default.abi,
            args: [],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: BigInt("19340180000000"),
        });

        console.log(`address from deployment is: ${l2EnshrinedTokenBridgeImplAddress}`);
        await waitTillCompleted(deployerAccount.client, l2EnshrinedTokenBridgeImplDepTxHash);
        console.log("✅ Logic Contract deployed at:", l2EnshrinedTokenBridgeImplDepTxHash);

        if (!l2EnshrinedTokenBridgeImplDepTxHash) {
            throw Error(`Invalid transaction output from deployContract call for L2EnshrinedTokenBridge Contract`);
        }

        if (!l2EnshrinedTokenBridgeImplAddress) {
            throw Error(`Invalid address output from deployContract call for L2EnshrinedTokenBridge Contract`);
        }

        console.log(`NilMessageTree contract deployed at address: ${l2EnshrinedTokenBridgeImplAddress} and with transactionHash: ${l2EnshrinedTokenBridgeImplDepTxHash}`);

        l2NetworkConfig.l2EnshrinedTokenBridgeConfig.l2EnshrinedTokenBridgeContracts.l2EnshrinedTokenBridgeImplementation = l2EnshrinedTokenBridgeImplAddress;

        const initData = encodeFunctionData({
            abi: L2EnshrinedTokenBridgeJson.default.abi,
            functionName: "initialize",
            args: [l2NetworkConfig.l2CommonConfig.owner, l2NetworkConfig.l2CommonConfig.admin,
            l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerContracts.l2BridgeMessengerProxy],
        });

        const { address: addressProxy, hash: hashProxy } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: TransparentUpgradeableProxy.default.bytecode,
            abi: TransparentUpgradeableProxy.default.abi,
            args: [l2EnshrinedTokenBridgeImplAddress, deployerAccount.address, initData],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: convertEthToWei(0.001),
        });
        await waitTillCompleted(deployerAccount.client, hashProxy);
        console.log("✅ Transparent Proxy Contract deployed at:", addressProxy);

        l2NetworkConfig.l2EnshrinedTokenBridgeConfig.l2EnshrinedTokenBridgeContracts.l2EnshrinedTokenBridgeProxy = addressProxy;

        console.log("Waiting 5 seconds...");
        await new Promise((res) => setTimeout(res, 5000));

        const fetchImplementationCall = encodeFunctionData({
            abi: TransparentUpgradeableProxy.default.abi,
            functionName: "fetchImplementation",
            args: [],
        });

        const fetchImplementationResult = await deployerAccount.client.call({
            to: addressProxy,
            data: fetchImplementationCall,
            from: deployerAccount.address,
        }, "latest");

        console.log(`L2EnshrinedTokenBridgeVaultProxy has fetch-implementation-result: ${JSON.stringify(fetchImplementationResult)}`);

        const proxyImplementationAddress = decodeFunctionResult({
            abi: TransparentUpgradeableProxy.default.abi,
            functionName: "fetchImplementation",
            data: fetchImplementationResult.data,
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

        console.log(`L2EnshrinedTokenBridgeProxy has admin-result: ${JSON.stringify(adminResult)}`);

        const proxyAdminAddress = decodeFunctionResult({
            abi: TransparentUpgradeableProxy.default.abi,
            functionName: "fetchAdmin",
            data: adminResult.data,
        }) as string;

        console.log("✅ ProxyAdmin Address:", proxyAdminAddress);

        l2NetworkConfig.l2EnshrinedTokenBridgeConfig.l2EnshrinedTokenBridgeContracts.proxyAdmin = proxyAdminAddress;

        // Save the updated config
        saveNilNetworkConfig(networkName, l2NetworkConfig);
    });
