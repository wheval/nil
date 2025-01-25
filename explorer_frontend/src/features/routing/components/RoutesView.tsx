import { ErrorPage } from "@nilfoundation/ui-kit";
import { createRoutesView } from "atomic-router-react";
import { AddressPage } from "../../../pages/address";
import { BlockPage } from "../../../pages/block";
import { ExplorerPage } from "../../../pages/explorer";
import { SandboxPage } from "../../../pages/sandbox";
import { TransactionPage } from "../../../pages/transaction";
import { addressRoute, addressTransactionsRoute } from "../routes/addressRoute";
import { blockDetailsRoute, blockRoute } from "../routes/blockRoute";
import { explorerRoute } from "../routes/explorerRoute";
import { sandboxRoute, sandboxWithHashRoute } from "../routes/sandboxRoute";
import { transactionRoute } from "../routes/transactionRoute";

export const RoutesView = createRoutesView({
  routes: [
    { route: explorerRoute, view: ExplorerPage },
    { route: transactionRoute, view: TransactionPage },
    { route: blockRoute, view: BlockPage },
    { route: blockDetailsRoute, view: BlockPage },
    {
      route: addressRoute,
      view: AddressPage,
    },
    {
      route: addressTransactionsRoute,
      view: AddressPage,
    },
    {
      route: sandboxRoute,
      view: SandboxPage,
    },
    {
      route: sandboxWithHashRoute,
      view: SandboxPage,
    },
  ],
  otherwise() {
    return (
      <ErrorPage
        redirectPath="/"
        errorCode={404}
        redirectTitle="Back to explorer"
        errorDescription="This page does not exist"
      />
    );
  },
});
