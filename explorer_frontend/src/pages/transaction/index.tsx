import loadable from "@loadable/component";
import { Spinner } from "@nilfoundation/ui-kit";

export const TransactionPage = loadable(
  async () => {
    const imported = await import("./TransactionPage");
    return imported.TransactionPage;
  },
  {
    fallback: <Spinner />,
  },
);
