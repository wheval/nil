import { createRoute } from "../utils/createRoute";

export const playgroundRoute = createRoute();
export const playgroundWithHashRoute = createRoute<{ snippetHash: string }>();
