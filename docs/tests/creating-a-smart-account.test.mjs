import { FAUCET_GLOBAL, NIL_GLOBAL, RPC_GLOBAL } from "./globals";

//startNilJSSmartAccountImports
import { HttpTransport, PublicClient, generateSmartAccount } from "@nilfoundation/niljs";
//endNilJSSmartAccountImports
import { NEW_SMART_ACCOUNT_PATTERN } from "./patterns";

const RPC_ENDPOINT = RPC_GLOBAL;
const FAUCET_ENDPOINT = FAUCET_GLOBAL;
const CONFIG_FILE_NAME = "tempConfigCreatingASmartAccount.ini";

const util = require("node:util");
const exec = util.promisify(require("node:child_process").exec);

const SALT = BigInt(Math.floor(Math.random() * 10000));

const CONFIG_FLAG = `--config ./tests/${CONFIG_FILE_NAME}`;

const CONFIG_COMMAND = `${NIL_GLOBAL} config init ${CONFIG_FLAG}`;
const RPC_COMMAND = `${NIL_GLOBAL} config set rpc_endpoint ${RPC_ENDPOINT} ${CONFIG_FLAG}`;
const FAUCET_COMMAND = `${NIL_GLOBAL} config set faucet_endpoint ${FAUCET_ENDPOINT} ${CONFIG_FLAG}`;
const KEYGEN_COMMAND = `${NIL_GLOBAL} keygen new ${CONFIG_FLAG}`;

//startSmartAccount
const SMART_ACCOUNT_CREATION_COMMAND = `${NIL_GLOBAL} smart-account new --salt ${SALT} ${CONFIG_FLAG}`;
//endSmartAccount

//startBalance
const SMART_ACCOUNT_BALANCE_COMMAND = `${NIL_GLOBAL} smart-account balance ${CONFIG_FLAG}`;
//endBalance

beforeAll(async () => {
  await exec(CONFIG_COMMAND);
  await exec(KEYGEN_COMMAND);
  await exec(RPC_COMMAND);
  await exec(FAUCET_COMMAND);
});

afterAll(async () => {
  await exec(`rm -rf ./tests/${CONFIG_FILE_NAME}`);
});

describe.sequential("initial CLI tests", () => {
  test.sequential("smart account creation command creates a smart account", async () => {
    await exec(KEYGEN_COMMAND);
    const { stdout, stderr } = await exec(SMART_ACCOUNT_CREATION_COMMAND);
    expect(stdout).toMatch(NEW_SMART_ACCOUNT_PATTERN);
  });

  test.sequential("smart account balance command returns balance", async () => {
    const pattern = /Smart account balance/;
    const { stdout, stderr } = await exec(SMART_ACCOUNT_BALANCE_COMMAND);
    expect(stdout).toMatch(pattern);
  });
});

describe.sequential("niljs test", () => {
  test.sequential(
    "niljs snippet can create and deploy a smart account",
    async () => {
      //startNilJSSmartAccountCreation

      const client = new PublicClient({
        transport: new HttpTransport({
          endpoint: RPC_ENDPOINT,
        }),
        shardId: 1,
      });

      const smartAccount = await generateSmartAccount({
        shardId: 1,
        rpcEndpoint: RPC_ENDPOINT,
        faucetEndpoint: FAUCET_ENDPOINT,
      });

      //endNilJSSmartAccountCreation
      expect(smartAccount.address).toBeDefined();
      const smartAccountCode = await client.getCode(smartAccount.address, "latest");
      expect(smartAccountCode).toBeDefined();
      expect(smartAccountCode.length).toBeGreaterThan(10);
    },
    50000,
  );
});
