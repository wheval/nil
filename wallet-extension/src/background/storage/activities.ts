export interface Activity {
  activityType: ActivityType;
  txHash: string;
  success: boolean;
  amount: string;
  token: string;
}

export enum ActivityType {
  SEND = "send",
  TOPUP = "top-up",
}

const STORAGE_KEY_PREFIX = "activities";

// Utility to get the storage key for a smartAccount
const getStorageKey = (smartAccountAddress: string): string => {
  return `${STORAGE_KEY_PREFIX}_${smartAccountAddress}`;
};

// Save an activity for a specific smartAccount
export const saveActivity = async (
  smartAccountAddress: string,
  activity: Activity,
): Promise<void> => {
  const storageKey = getStorageKey(smartAccountAddress);
  const result = await chrome.storage.local.get(storageKey);
  const existingActivities: Activity[] = result[storageKey] || [];
  const updatedActivities = [...existingActivities, activity];
  await chrome.storage.local.set({ [storageKey]: updatedActivities });
};

// Get all activities for a specific smartAccount
export const getActivities = async (smartAccountAddress: string): Promise<Activity[]> => {
  const storageKey = getStorageKey(smartAccountAddress);
  const result = await chrome.storage.local.get(storageKey);
  return result[storageKey] || [];
};
