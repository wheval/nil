import { createEvent, createStore } from "../store.ts";

// Endpoint
export const setEndpoint = createEvent<string>();
export const resetEndpoint = createEvent();


// Store
export const $endpoint = createStore<string>("")
  .on(setEndpoint, (_, newEndpoint) => newEndpoint)
  .reset(resetEndpoint);

$endpoint.watch((endpoint) => {
  console.log("Endpoint updated:", endpoint);
});
