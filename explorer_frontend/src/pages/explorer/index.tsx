import loadable from "@loadable/component";
import { Spinner } from "@nilfoundation/ui-kit";

export const ExplorerPage = loadable(
  async () => {
    const imported = import("./ExplorerPage");
    return (await imported).ExplorerPage;
  },
  {
    fallback: <Spinner />,
  },
);
