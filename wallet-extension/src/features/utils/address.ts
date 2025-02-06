import type { Hex } from "@nilfoundation/niljs";

// Formats a Hex address into a shortened form (e.g., 0x123...456)
export function formatAddress(address: Hex): string {
  // Ensure the address starts with '0x' and is at least long enough to format
  if (!address.startsWith("0x") || address.length <= 10) {
    throw new Error("Invalid Hex address provided.");
  }

  const firstPart = address.slice(0, 9);
  const lastPart = address.slice(-4);
  return `${firstPart}...${lastPart}`;
}

// Generate a random salt between 1 and 10000
export const generateRandomSalt = (): bigint => {
  return BigInt(Math.floor(Math.random() * 10000) + 1);
};

// Generate a random shard number between 1 and VITE_NUMBER_SHARDS
export const generateRandomShard = (): number => {
  const numberOfShards = Number(import.meta.env.VITE_NUMBER_SHARDS);
  if (!numberOfShards || numberOfShards < 1) {
    throw new Error("Environment variable VITE_NUMBER_SHARDS must be a positive integer.");
  }

  return Math.floor(Math.random() * numberOfShards) + 1;
};

// Returns the last 2 bytes of a Hex address
export function getLast2Bytes(address: Hex): string {
  // Ensure the address starts with '0x' and is at least 10 characters long
  if (!address.startsWith("0x") || address.length < 10) {
    throw new Error("Invalid Hex address provided.");
  }

  // Extract the last 2 bytes (4 hex characters)
  const last2Bytes = address.slice(-4);
  return `0x...${last2Bytes}`;
}
