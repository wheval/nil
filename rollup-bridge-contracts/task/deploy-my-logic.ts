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

        console.log(`adminResult queried is: ${JSON.stringify(adminResult)}`);

        const proxyAdminAddress = decodeFunctionResult({
            abi: TransparentUpgradeableProxy.default.abi,
            functionName: "fetchAdmin",
            data: adminResult.data,
        }) as string;

        console.log("✅ ProxyAdmin Address:", proxyAdminAddress);

        const owner = encodeFunctionData({
            abi: ProxyAdmin.default.abi,
            functionName: "owner",
            args: [],
        })

        const ownerResult = await deployerAccount.client.call({
            to: proxyAdminAddress as `0x${string}`,
            data: owner,
            from: deployerAccount.address,
        }, "latest");

        const proxyAdminOwner = decodeFunctionResult({
            abi: ProxyAdmin.default.abi,
            functionName: "owner",
            data: ownerResult.data,
        }) as string;

        console.log("✅ ProxyAdmin Owner:", proxyAdminOwner);

        const getValueData = encodeFunctionData({
            abi: MyLogicJson.default.abi,
            functionName: "value",
            args: [],
        });
        const getValueCall = await deployerAccount.client.call({
            to: addressProxy,
            from: deployerAccount.address,
            data: getValueData,
        }, "latest");

        const getValue = decodeFunctionResult({
            abi: MyLogicJson.default.abi,
            functionName: "value",
            data: getValueCall.data,
        });
        console.log("✅ Current value in Logic contract:", getValue);
    });
