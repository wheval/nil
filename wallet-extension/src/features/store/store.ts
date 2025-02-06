import { createDomain } from "effector";

export const accountConnectorDomain = createDomain("account-connector");
export const createStore = accountConnectorDomain.createStore.bind(accountConnectorDomain);
export const createEvent = accountConnectorDomain.createEvent.bind(accountConnectorDomain);
export const createEffect = accountConnectorDomain.createEffect.bind(accountConnectorDomain);
