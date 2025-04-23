import type { FC, ReactNode } from "react";
import { Client as Styletron } from "styletron-engine-atomic";
import { BaseProvider } from "baseui";
import { Provider } from "styletron-react";
import { createTheme } from "@nilfoundation/ui-kit";
import { router } from "../../src/features/routing/routes/routes";
import { RouterProvider } from "atomic-router-react";

type TestLayoutProps = {
  children: ReactNode;
};

const engine = new Styletron();
const { theme } = createTheme(engine);

export const TestsLayout: FC<TestLayoutProps> = ({ children }) => {
  return (
    <Provider value={engine}>
      <BaseProvider theme={theme}>
        <RouterProvider router={router}>{children}</RouterProvider>
      </BaseProvider>
    </Provider>
  );
};
