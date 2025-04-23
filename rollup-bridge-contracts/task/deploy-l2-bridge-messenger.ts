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


        const { tx, address } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: L2BridgeMessengerJson.bytecode,
            abi: L2BridgeMessengerJson.abi,
            args: [deployerAccount.address],
            salt: BigInt(Math.floor(Math.random() * 10000)),
            feeCredit: BigInt("19340180000000"),
        });


        if (!tx) {
            throw Error(`Invalid transaction output from deployContract call for NilMessageTree Contract`);
        }

        if (!address) {
            throw Error(`Invalid address output from deployContract call for NilMessageTree Contract`);
        }

        const implDeploymentTranasction: Transaction = tx;

        console.log(`tx from deployment is: ${JSON.stringify(implDeploymentTranasction)}`);


        await waitTillCompleted(deployerAccount.client, implDeploymentTranasction.hash);
        console.log("✅ Logic Contract deployed at:", address);
    });