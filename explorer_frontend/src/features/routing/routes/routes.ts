import { createHistoryRouter, createRoute } from "atomic-router";
import { createBrowserHistory } from "history";
import { addressRoute, addressTransactionsRoute } from "./addressRoute";
import { blockDetailsRoute, blockRoute } from "./blockRoute";
import { explorerRoute } from "./explorerRoute";
import { sandboxRoute, sandboxWithHashRoute } from "./sandboxRoute";
import { transactionRoute } from "./transactionRoute";

export const notFoundRoute = createRoute();

export const routes = [
  {
    path: "/",
    route: explorerRoute,
  },
  {
    path: "/tx/:hash",
    route: transactionRoute,
  },
  {
    path: "/block/:shard/:id",
    route: blockRoute,
  },
  {
    path: "/block/:shard/:id/:details",
    route: blockDetailsRoute,
  },
  {
    path: "/address/:address",
    route: addressRoute,
  },
  {
    path: "/address/:address/transactions",
    route: addressTransactionsRoute,
  },
  {
    path: "/sandbox",
    route: sandboxRoute,
  },
  {
    path: "/sandbox/:snippetHash",
    route: sandboxWithHashRoute,
  },
];

export const router = createHistoryRouter({
  routes,
  notFoundRoute,
});

export const history = createBrowserHistory();

router.setHistory(history);
