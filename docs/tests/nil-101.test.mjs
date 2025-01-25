import TestHelper from "./TestHelper";
import { CALLER_COMPILATION_COMMAND, COUNTER_COMPILATION_COMMAND } from "./compilationCommands";
import { NIL_GLOBAL } from "./globals";
import {
  ADDRESS_PATTERN,
  CONTRACT_ADDRESS_PATTERN,
  FAUCET_PATTERN,
  NEW_SMART_ACCOUNT_PATTERN,
  PRIVATE_KEY_PATTERN,
  RPC_PATTERN,
  SMART_ACCOUNT_ADDRESS_PATTERN,
  SMART_ACCOUNT_BALANCE_PATTERN,
  TOKEN_PATTERN,
  TRANSACTION_HASH_PATTERN,
} from "./patterns";

const util = require("node:util");
const exec = util.promisify(require("node:child_process").exec);

let SALT = BigInt(Math.floor(Math.random() * 10000));

const CONFIG_FILE_NAME = "./tests/tempConfigNil101.ini";

const CONFIG_FLAG = `--config ${CONFIG_FILE_NAME}`;

let TEST_COMMANDS;
let COUNTER_ADDRESS;
let CALLER_ADDRESS;
let NEW_SMART_ACCOUNT_ADDRESS;

beforeAll(async () => {
  const testHelper = new TestHelper({ configFileName: CONFIG_FILE_NAME });
  TEST_COMMANDS = testHelper.createCLICommandsMap(SALT);
  await exec(TEST_COMMANDS.CONFIG_COMMAND);
});

afterAll(async () => {
  await exec(`rm -rf ${CONFIG_FILE_NAME}`);
});

describe.sequential("initial smart account setup tests", () => {
  test.sequential("keygen generation works via CLI", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.KEYGEN_COMMAND);
    expect(stdout).toMatch(PRIVATE_KEY_PATTERN);
  });

  test.sequential("endpoint command should set the endpoint", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.RPC_COMMAND);
    expect(stderr).toMatch(RPC_PATTERN);
  });

  test.sequential("faucet_endpoint command should set the faucet endpoint", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.FAUCET_COMMAND);
    expect(stderr).toMatch(FAUCET_PATTERN);
  });

  test.sequential("smart account creation command creates a smart account", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.SMART_ACCOUNT_CREATION_COMMAND);
    expect(stdout).toMatch(NEW_SMART_ACCOUNT_PATTERN);
  });

  test.sequential("smart account top-up command tops up the smart account", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.SMART_ACCOUNT_TOP_UP_COMMAND);
    expect(stdout).toMatch(SMART_ACCOUNT_BALANCE_PATTERN);
  });
});

describe.sequential("incrementer tests", () => {
  test.sequential("smart account info command supplies info", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.SMART_ACCOUNT_INFO_COMMAND);
    expect(stdout).toMatch(SMART_ACCOUNT_ADDRESS_PATTERN);
  });

  test.sequential("deploy of incrementer works successfully", async () => {
    await exec(COUNTER_COMPILATION_COMMAND);
    const { stdout, stderr } = await exec(TEST_COMMANDS.COUNTER_DEPLOYMENT_COMMAND);
    expect(stdout).toMatch(CONTRACT_ADDRESS_PATTERN);
    const addressMatches = stdout.match(ADDRESS_PATTERN);
    COUNTER_ADDRESS = addressMatches.length > 1 ? addressMatches[1] : null;
  });

  test.sequential("execution of increment produces a transaction", async () => {
    //startIncrement
    const COUNTER_INCREMENT_COMMAND = `${NIL_GLOBAL} smart-account send-transaction ${COUNTER_ADDRESS} increment --abi ./tests/Counter/Counter.abi ${CONFIG_FLAG}`;
    //endIncrement
    const { stdout, stderr } = await exec(COUNTER_INCREMENT_COMMAND);
    expect(stdout).toMatch(TRANSACTION_HASH_PATTERN);
  });

  test.sequential("call to incrementer returns the correct value", async () => {
    //start_CallToIncrementer
    const COUNTER_CALL_READONLY_COMMAND = `${NIL_GLOBAL} contract call-readonly ${COUNTER_ADDRESS} getValue --abi ./tests/Counter/Counter.abi ${CONFIG_FLAG}`;
    //end_CallToIncrementer
    const { stdout, stderr } = await exec(COUNTER_CALL_READONLY_COMMAND);

    const normalize = (str) => str.replace(/\r\n/g, "\n").trim();

    const expectedOutput = "Success, result:\nuint256: 1";
    const receivedOutput = normalize(stdout);

    expect(receivedOutput).toBe(expectedOutput);
  });
});

describe.sequential("caller tests", () => {
  beforeEach(() => {
    SALT = BigInt(Math.floor(Math.random() * 10000));
  });
  test.sequential("deploy of caller works successfully", async () => {
    await exec(CALLER_COMPILATION_COMMAND);
    const { stdout, stderr } = await exec(TEST_COMMANDS.CALLER_DEPLOYMENT_COMMAND);
    const addressMatches = stdout.match(ADDRESS_PATTERN);
    CALLER_ADDRESS = addressMatches && addressMatches.length > 0 ? addressMatches[1] : null;
    expect(CALLER_ADDRESS).not.toBeNull();
  });

  test.sequential("caller can call incrementer successfully", async () => {
    //start_SendTokensToCaller
    const SEND_TOKENS_COMMAND = `${NIL_GLOBAL} smart-account send-tokens ${CALLER_ADDRESS} 3000000 ${CONFIG_FLAG}`;
    //end_SendTokensToCaller

    //startTransactionFromCallerToIncrementer
    const SEND_FROM_CALLER_COMMAND = `${NIL_GLOBAL} smart-account send-transaction ${CALLER_ADDRESS} call ${COUNTER_ADDRESS} --abi ./tests/Caller/Caller.abi ${CONFIG_FLAG}`;
    //endTransactionFromCallerToIncrementer

    await exec(SEND_TOKENS_COMMAND);
    const { stdout, stderr } = await exec(SEND_FROM_CALLER_COMMAND);
    expect(stdout).toMatch(TRANSACTION_HASH_PATTERN);

    const COUNTER_CALL_READONLY_COMMAND_POST_CALLER = `${NIL_GLOBAL} contract call-readonly ${COUNTER_ADDRESS} getValue --abi ./tests/Counter/Counter.abi ${CONFIG_FLAG}`;

    let stdoutCall;
    let stderrCall;

    try {
      for (let attempt = 0; attempt < 5; attempt++) {
        ({ stdout: stdoutCall, stderr: stderrCall } = await exec(
          COUNTER_CALL_READONLY_COMMAND_POST_CALLER,
        ));

        if (stdoutCall) {
          break;
        }

        console.log(`Attempt ${attempt + 1}: Retrying after a short delay...`);
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }

      if (!stdoutCall) {
        throw new Error("Failed to get output from the contract call after multiple attempts.");
      }

      const normalize = (str) => str.replace(/\r\n/g, "\n").trim();

      const expectedOutput = "Success, result:\nuint256: 2";
      const receivedOutput = normalize(stdoutCall);

      expect(receivedOutput).toBe(expectedOutput);
    } catch (error) {
      console.error("Error during the contract call:", error);
      if (stderrCall) {
        console.error("stderrCall:", stderrCall);
      }
      throw error;
    }
  });
});

describe.sequential("tokens tests", () => {
  test.sequential("a new smart account is created successfully", async () => {
    const { stdout, stderr } = await exec(TEST_COMMANDS.SMART_ACCOUNT_CREATION_COMMAND_WITH_SALT);
    expect(stdout).toMatch(SMART_ACCOUNT_ADDRESS_PATTERN);
    const addressMatches = stdout.match(SMART_ACCOUNT_ADDRESS_PATTERN);
    NEW_SMART_ACCOUNT_ADDRESS =
      addressMatches && addressMatches.length > 0 ? addressMatches[0] : null;
  });

  test.sequential("a new token is created and withdrawn successfully", async () => {
    //startMintToken
    const MINT_TOKEN_COMMAND = `${NIL_GLOBAL} minter create-token ${NEW_SMART_ACCOUNT_ADDRESS} 5000 new-token ${CONFIG_FLAG}`;
    //endMintToken

    await exec(MINT_TOKEN_COMMAND);

    //startTokensCheck
    const TOKENS_COMMAND = `${NIL_GLOBAL} contract tokens ${NEW_SMART_ACCOUNT_ADDRESS} ${CONFIG_FLAG}`;
    //endTokensCheck

    const { stdout, stderr } = await exec(TOKENS_COMMAND);
    expect(stdout).toMatch(TOKEN_PATTERN);
  });
});
