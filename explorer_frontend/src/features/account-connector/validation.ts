import { Token } from "../tokens";

export type ValidationResult = {
  error: string;
  isValid: boolean;
};

// Validates if the provided RPC url matches the expected format
export const validateRpcUrl = (rpcUrl: string): ValidationResult => {
  const RPC_REGEX = /^https:\/\/api\.devnet\.nil\.foundation\/api\/.+\/.+$/;
  if (RPC_REGEX.test(rpcUrl)) {
    return { isValid: true, error: "" };
  }
  return { isValid: false, error: "Invalid RPC rpcUrl format" };
};

export const MAX_AMOUNT_NIL = 1;
export const MIN_AMOUNT_NIL = 0.0001;
export const MAX_AMOUNT_OTHER = 100;
export const MIN_AMOUNT_OTHER = 1;

export function validateAmount(amount: string, selectedCurrency: string): string | null {
  if (!amount.trim()) return "Please enter an amount";

  const numericAmount = Number(amount);
  if (Number.isNaN(numericAmount)) return "Invalid input. Please enter a valid number";

  return selectedCurrency === Token.NIL
    ? validateNilAmount(numericAmount)
    : validateOtherCurrencyAmount(numericAmount, selectedCurrency);
}

function validateNilAmount(amount: number): string | null {
  if (amount < MIN_AMOUNT_NIL) return `Minimum allowed amount is ${MIN_AMOUNT_NIL} NIL`;
  if (amount > MAX_AMOUNT_NIL) return `Maximum allowed amount is ${MAX_AMOUNT_NIL} NIL`;
  return null;
}

function validateOtherCurrencyAmount(amount: number, currency: string): string | null {
  if (amount < MIN_AMOUNT_OTHER) return `Minimum allowed amount is ${MIN_AMOUNT_OTHER} ${currency}`;
  if (amount > MAX_AMOUNT_OTHER) return `Maximum allowed amount is ${MAX_AMOUNT_OTHER} ${currency}`;
  if (!Number.isInteger(amount)) return `${currency} does not support decimal values`;
  return null;
}
