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

// npx hardhat deploy-nil-message-tree  --networkname local
task("deploy-nil-message-tree", "Deploys NilMessageTree contract on Nil Chain")
    .addParam("networkname", "The network to use") // Mandatory parameter
    .setAction(async (taskArgs) => {
        const deployerAccount = await loadNilSmartAccount();

        if (!deployerAccount) {
            throw Error(`Invalid Deployer SmartAccount`);
        }

        const balance = await deployerAccount.getBalance();

        console.log(`smart-contract${deployerAccount.address} is on shard: ${deployerAccount.shardId} with balance: ${balance}`);

        if (!(balance > BigInt(0))) {
            throw Error(`Insufficient or Zero balance for smart-account: ${deployerAccount.address}`);
        }

        const NilMessageTreeJson = require("../artifacts/contracts/common/NilMessageTree.sol/NilMessageTree.json");

        if (!NilMessageTreeJson || !NilMessageTreeJson.abi || !NilMessageTreeJson.bytecode) {
            throw Error(`Invalid NilMessageTree ABI`);
        }

        const { tx, address } = await deployerAccount.deployContract({
            shardId: 1,
            bytecode: NilMessageTreeJson.bytecode,
            abi: NilMessageTreeJson.abi,
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

        const transactionData: Transaction = tx;

        console.log(`tx from deployment is: ${JSON.stringify(transactionData)}`);

        await waitTillCompleted(deployerAccount.client, transactionData.hash);

        console.log("✅ Logic Contract deployed at:", address);

        console.log("NilMessageTree contract deployed at address: " + address);
    });