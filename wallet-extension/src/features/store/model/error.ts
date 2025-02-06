import { createEvent, createStore } from "../store.ts";

// Global error store
export const $globalError = createStore<string | null>(null);

// Event to set a global error
export const setGlobalError = createEvent<string>();

// Event to reset the global error
export const resetGlobalError = createEvent();

// Update the error store when a new error is set
$globalError.on(setGlobalError, (_, error) => error);

// Reset the error store when the reset event is triggered
$globalError.reset(resetGlobalError);
