import { COMETA_GLOBAL, FAUCET_GLOBAL, NIL_GLOBAL, RPC_GLOBAL } from "./globals";

const SALT = BigInt(Math.floor(Math.random() * 10000));

const RPC_ENDPOINT = RPC_GLOBAL;

//startConfigCommand
const CONFIG_COMMAND = `${NIL_GLOBAL} config init`;
//endConfigCommand

//startKeygen
const KEYGEN_COMMAND = `${NIL_GLOBAL} keygen new`;
//endKeygen

//startEndpoint
const RPC_COMMAND = `${NIL_GLOBAL} config set rpc_endpoint ${RPC_ENDPOINT}`;
//endEndpoint

//startSmartAccount
const SMART_ACCOUNT_CREATION_COMMAND = `${NIL_GLOBAL} smart-account new`;
//endSmartAccount

//startTopup
const SMART_ACCOUNT_TOP_UP_COMMAND = `${NIL_GLOBAL} smart-account top-up 1000000`;
//endTopup

//startInfo
const SMART_ACCOUNT_INFO_COMMAND = `${NIL_GLOBAL} smart-account info`;
//endInfo

//startSaltSmartAccountCreation
const SMART_ACCOUNT_CREATION_COMMAND_WITH_SALT = `${NIL_GLOBAL} smart-account new --salt ${SALT}`;
//endSaltSmartAccountCreation

//startLatestBlock
const RETRIEVE_LATEST_BLOCK_COMMAND = `${NIL_GLOBAL} block latest --shard-id 1`;
//endLatestBlock

//startCounterDeploymentCommand
const COUNTER_DEPLOYMENT_COMMAND = `${NIL_GLOBAL} smart-account deploy ./tests/Counter/Counter.bin --salt ${SALT}`;
//endCounterDeploymentCommand

//startCallerDeploy
const CALLER_DEPLOYMENT_COMMAND = `${NIL_GLOBAL} smart-account deploy ./tests/Caller/Caller.bin --shard-id 2 --salt ${SALT}`;
//endCallerDeploy

const AWAITER_DEPLOYMENT_COMMAND = `${NIL_GLOBAL} smart-account deploy ./tests/Awaiter/Awaiter.bin --abi ./tests/Awaiter/Awaiter.abi --shard-id 2 --salt ${SALT}`;

//startCounterBugDeploymentCommand
const COUNTER_BUG_DEPLOYMENT_COMMAND = `${NIL_GLOBAL} smart-account deploy ./tests/CounterBug/CounterBug.bin --abi ./tests/CounterBug/CounterBug.abi --shard-id 2 --salt ${SALT}`;
//endCounterBugDeploymentCommand

//startCometaEndpointCommand
const COMETA_ENDPOINT_COMMAND = `${NIL_GLOBAL} config set cometa_endpoint ${COMETA_GLOBAL}`;
//endCometaEndpointCommand

//startCometaCommand
const COUNTER_BUG_COMETA_COMMAND = `${NIL_GLOBAL} smart-account deploy --compile-input ./tests/counter.json --salt ${SALT}`;
//endCometaCommand

//startFaucetEndpointCommand
const FAUCET_ENDPOINT_COMMAND = `${NIL_GLOBAL} config set faucet_endpoint ${FAUCET_GLOBAL}`;
//endFaucetEndpointCommand

const COMMANDS = {
  CONFIG_COMMAND: CONFIG_COMMAND,
  KEYGEN_COMMAND: KEYGEN_COMMAND,
  RPC_COMMAND: RPC_COMMAND,
  FAUCET_COMMAND: FAUCET_ENDPOINT_COMMAND,
  COMETA_COMMAND: COMETA_ENDPOINT_COMMAND,
  SMART_ACCOUNT_CREATION_COMMAND: SMART_ACCOUNT_CREATION_COMMAND,
  SMART_ACCOUNT_TOP_UP_COMMAND: SMART_ACCOUNT_TOP_UP_COMMAND,
  SMART_ACCOUNT_INFO_COMMAND: SMART_ACCOUNT_INFO_COMMAND,
  SMART_ACCOUNT_CREATION_COMMAND_WITH_SALT: SMART_ACCOUNT_CREATION_COMMAND_WITH_SALT,
  RETRIEVE_LATEST_BLOCK_COMMAND: RETRIEVE_LATEST_BLOCK_COMMAND,
  COUNTER_DEPLOYMENT_COMMAND: COUNTER_DEPLOYMENT_COMMAND,
  CALLER_DEPLOYMENT_COMMAND: CALLER_DEPLOYMENT_COMMAND,
  AWAITER_DEPLOYMENT_COMMAND: AWAITER_DEPLOYMENT_COMMAND,
  COUNTER_BUG_COMETA_COMMAND: COUNTER_BUG_COMETA_COMMAND,
  COUNTER_BUG_DEPLOYMENT_COMMAND: COUNTER_BUG_DEPLOYMENT_COMMAND,
};

export default COMMANDS;
