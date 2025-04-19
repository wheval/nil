import type { Hex } from "@nilfoundation/niljs";
import type { Transaction } from "@nilfoundation/niljs";
import { Args, Flags } from "@oclif/core";
import type { Abi } from "abitype";
import { BaseCommand } from "../../base.js";
import { readJsonFile } from "../../common/utils.js";
import { bigintFlag, hexArg, tokenFlag } from "../../types.js";

export default class SmartAccountSendTransaction extends BaseCommand {
  static override summary = "Send a transaction to a smart contract via the smart account";
  static override description =
    "Send a transaction to the smart contract with the specified bytecode or command via the smart account";

  static flags = {
    abiPath: Flags.string({
      char: "a",
      description: "The path to the ABI file",
      required: true,
    }),
    amount: bigintFlag({
      char: "m",
      description: "The amount of default tokens to send",
      required: false,
    }),
    noWait: Flags.boolean({
      char: "n",
      description: "Define whether the command should wait for the receipt",
      default: false,
    }),
    feeCredit: bigintFlag({
      char: "f",
      description: "The fee credit for transaction processing",
      required: false,
    }),
    tokens: tokenFlag({
      char: "c",
      description:
        "The custom tokens to transfer in as a map 'tokenId=amount', can be set multiple times",
      multiple: true,
      required: false,
    }),
  };

  static args = {
    address: hexArg({
      name: "address",
      required: true,
      description: "The address of the smart contract",
    }),
    bytecodeOrMethod: Args.string({
      name: "bytecodeOrMethod",
      required: true,
      description: "The bytecode or method to send",
    }),
    args: Args.string({
      name: "args",
      required: false,
      description: "Arguments for the method",
      multiple: true,
    }),
  };

  static override examples = ["<%= config.bin %> <%= command.id %>"];

  public async run(): Promise<Hex> {
    const { flags, args } = await this.parse(SmartAccountSendTransaction);

    const { smartAccount } = await this.setupSmartAccount();

    const address = args.address as Hex;
    let abi: Abi;
    try {
      abi = readJsonFile<Abi>(flags.abiPath);
    } catch (e) {
      this.error(`Invalid ABI file: ${e}`);
    }

    let tx: Transaction;

    if (args.bytecodeOrMethod.startsWith("0x")) {
      const data = args.bytecodeOrMethod as Hex;
      tx = await smartAccount.sendTransaction({
        to: address,
        value: flags.amount ?? 0n,
        feeCredit: flags.feeCredit ?? 0n,
        tokens: flags.tokens ?? [],
        data: data,
      });
    } else {
      tx = await smartAccount.sendTransaction({
        to: address,
        value: flags.amount ?? 0n,
        feeCredit: flags.feeCredit ?? 0n,
        args: args.args?.split(" ") ?? [],
        abi: abi,
        functionName: args.bytecodeOrMethod,
        tokens: flags.tokens ?? [],
      });
    }
    if (flags.quiet) {
      this.log(tx.hash);
    } else {
      this.log(`Transaction hash: ${tx.hash}`);
    }

    if (flags.noWait) {
      return tx.hash;
    }

    this.info("Waiting for the transaction to be processed...");
    await tx.wait();
    this.info("Transaction successfully processed");

    return tx.hash;
  }
}
