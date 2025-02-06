import type { Hex, SmartAccountV1 } from "@nilfoundation/niljs";
import { createEffect, createEvent, createStore, sample } from "effector";
import { type Activity, getActivities, saveActivity } from "../../../background/storage";

export const $activities = createStore<Activity[]>([]);

// Events
export const initializeActivities = createEvent<SmartAccountV1>();
export const addActivity = createEvent<{ smartAccountAddress: Hex; activity: Activity }>();

// Effects
export const fetchActivitiesFx = createEffect<Hex, Activity[], Error>(
  async (smartAccountAddress) => {
    try {
      return await getActivities(smartAccountAddress);
    } catch (error) {
      console.error("Failed to fetch activities:", error);
      throw error;
    }
  },
);

export const saveActivityFx = createEffect<
  { smartAccountAddress: Hex; activity: Activity },
  void,
  Error
>(async ({ smartAccountAddress, activity }) => {
  try {
    await saveActivity(smartAccountAddress, activity);
  } catch (error) {
    console.error("Failed to save activity:", error);
    throw error;
  }
});

// Store
$activities.on(fetchActivitiesFx.doneData, (_, activities) => activities);
$activities.on(addActivity, (state, { activity }) => [...state, activity]);

// Automatically fetch activities when `initializeActivities` is triggered
sample({
  source: initializeActivities,
  fn: (smartAccount) => smartAccount.address,
  target: fetchActivitiesFx,
});

// Automatically save activity when `addActivity` is triggered
sample({
  source: addActivity,
  fn: ({ smartAccountAddress, activity }) => ({ smartAccountAddress, activity }),
  target: saveActivityFx,
});

// Watchers for debugging
$activities.watch((activities) => {
  console.log("Updated activities:", activities);
});

export const $latestActivity = createStore<Activity | null>(null);

// Event to clear latest activity
export const clearLatestActivity = createEvent();

// Update latestActivity when a new activity is added
// Store updates
$latestActivity.on(addActivity, (_, { activity }) => {
  // When a new activity is added, schedule a clear event
  scheduleAutoClear();
  return activity;
});

// Clear latestActivity when `clearLatestActivity` is called
$latestActivity.on(clearLatestActivity, () => null);

// Watcher for debugging
$latestActivity.watch((activity) => {
  console.log("Latest activity:", activity);
});

// Function to automatically clear the latest activity after 5 seconds
let timeoutId: ReturnType<typeof setTimeout> | null = null;
function scheduleAutoClear() {
  // Reset if another activity comes in
  if (timeoutId) clearTimeout(timeoutId);

  timeoutId = setTimeout(() => {
    clearLatestActivity();
    timeoutId = null;
  }, 5000);
}
