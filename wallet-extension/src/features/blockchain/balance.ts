import type { SmartAccountV1 } from "@nilfoundation/niljs";

// Fetch smartAccount balance
export async function fetchBalance(smartAccount: SmartAccountV1): Promise<bigint> {
  try {
    return await smartAccount.getBalance();
  } catch (error) {
    console.error("Error fetching smartAccount balance:", error);
    throw new Error("Failed to fetch smartAccount balance");
  }
}

// Fetch smartAccount currencies
export async function fetchSmartAccountCurrencies(
  smartAccount: SmartAccountV1,
): Promise<Record<string, bigint>> {
  try {
    return await smartAccount.client.getTokens(smartAccount.address, "latest");
  } catch (error) {
    console.error("Error fetching smartAccount currencies:", error);
    throw new Error("Failed to fetch smartAccount currencies");
  }
}
