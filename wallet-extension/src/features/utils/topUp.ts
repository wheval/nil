import { convertEthToWei } from "@nilfoundation/niljs";
import { TokenNames } from "../components/token";

export const MAX_AMOUNT_NIL = 1;
export const MIN_AMOUNT_NIL = 0.0001;
export const MAX_AMOUNT_OTHER = 100;
export const MIN_AMOUNT_OTHER = 1;

export function validateAmount(amount: string, selectedToken: string): string | null {
  if (!amount.trim()) return "Please enter an amount";

  const numericAmount = Number(amount);
  if (Number.isNaN(numericAmount)) return "Invalid input. Please enter a valid number";

  return selectedToken === TokenNames.NIL
    ? validateNilAmount(numericAmount)
    : validateOtherTokenAmount(numericAmount, selectedToken);
}

function validateNilAmount(amount: number): string | null {
  if (amount < MIN_AMOUNT_NIL) return `Minimum allowed amount is ${MIN_AMOUNT_NIL} NIL`;
  if (amount > MAX_AMOUNT_NIL) return `Maximum allowed amount is ${MAX_AMOUNT_NIL} NIL`;
  return null;
}

function validateOtherTokenAmount(amount: number, token: string): string | null {
  if (amount < MIN_AMOUNT_OTHER) return `Minimum allowed amount is ${MIN_AMOUNT_OTHER} ${token}`;
  if (amount > MAX_AMOUNT_OTHER) return `Maximum allowed amount is ${MAX_AMOUNT_OTHER} ${token}`;
  if (!Number.isInteger(amount)) return `${token} does not support decimal values`;
  return null;
}

export function convertAmount(amount: string, selectedToken: string): bigint {
  return selectedToken === TokenNames.NIL ? convertEthToWei(Number(amount)) : BigInt(amount);
}

export function getQuickAmounts(selectedToken: string): number[] {
  return selectedToken === TokenNames.NIL ? [0.0001, 0.003, 0.05] : [1, 5, 10];
}
