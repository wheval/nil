export type ValidationResult = {
  error: string;
  isValid: boolean;
};

// Validates if the provided RPC endpoint matches the expected format
export const validateRpcEndpoint = (rpcEndpoint: string): ValidationResult => {
  const RPC_REGEX = /^https:\/\/api\.devnet\.nil\.foundation\/api\/.+\/.+$/;
  if (RPC_REGEX.test(rpcEndpoint)) {
    return { isValid: true, error: "" };
  }
  return { isValid: false, error: "Invalid RPC endpoint format" };
};

// Validates if the provided smartAccount address is valid
export const validateSmartAccountAddress = (
  smartAccountAddress: string,
  address: string,
): ValidationResult => {
  const shardNumber = import.meta.env.VITE_NUMBER_SHARDS;
  const isValidLength = smartAccountAddress.length === 42;
  const isHex = /^0x[a-fA-F0-9]{40}$/.test(smartAccountAddress);
  const validPrefixes = Array.from(
    { length: shardNumber },
    (_, i) => `0x${(i + 1).toString().padStart(4, "0")}`,
  );
  const hasValidPrefix = validPrefixes.some((prefix) => smartAccountAddress.startsWith(prefix));

  if (isValidLength && isHex && hasValidPrefix && smartAccountAddress !== address) {
    return { isValid: true, error: "" };
  }

  if (smartAccountAddress.toLowerCase() === address.toLowerCase()) {
    return { isValid: false, error: "Recipient address cannot be your own smartAccount address" };
  }

  return { isValid: false, error: "Invalid smartAccount address" };
};
