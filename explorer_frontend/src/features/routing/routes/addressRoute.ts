import { createRoute } from "../utils/createRoute";

export const addressRoute = createRoute<{ address: string }>();
export const addressTransactionsRoute = createRoute<{ address: string }>();
