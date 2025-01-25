export const SMART_ACCOUNT_ADDRESS_PATTERN = /0x[a-fA-F0-9]{40}/;
export const ADDRESS_PATTERN = /0x[a-fA-F0-9]{40}/g;
export const CONTRACT_ADDRESS_PATTERN = /Contract address:/;
export const PUBKEY_PATTERN = /Public key:\s(0x[a-fA-F0-9]+)/;

export const RETAILER_COMPILATION_PATTERN = /Function state mutability/;
export const MANUFACTURER_COMPILATION_PATTERN = /Compiler run successful/;

export const CREATED_TOKEN_PATTERN = /Created Token ID:\s(0x[a-fA-F0-9]+)/;
export const TOKEN_PATTERN = /5000/;

export const SUCCESSFUL_EXECUTION_PATTERN = /Compiler run successful/;
export const PREV_BLOCK_PATTERN = /PrevBlock/;
export const HASH_PATTERN = /0x[a-fA-F0-9]{64}/g;
export const PRIVATE_KEY_PATTERN = /\bPrivate key: [a-f0-9]{64}\b/;
export const RPC_PATTERN = /Set "rpc_endpoint" to /;
export const FAUCET_PATTERN = /Set "faucet_endpoint" to /;
export const NEW_SMART_ACCOUNT_PATTERN = /New smart account address/;

export const SMART_ACCOUNT_BALANCE_PATTERN = /Smart Account balance/;
export const TRANSACTION_HASH_PATTERN = /Transaction hash:/;

export const ESCROW_SUCCESSFUL_PATTERN = /Function state mutability can be restricted to pure/;

export const COUNTER_BUG_DEBUG_PATTERN = /Transaction/;

export const SERVER_RUNNING_PATTERN = /Server running at/;
