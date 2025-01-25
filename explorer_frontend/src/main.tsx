import "./init";
import { createTheme } from "@nilfoundation/ui-kit";
import { BaseProvider } from "baseui";
import { createRoot } from "react-dom/client";
import { Client as Styletron } from "styletron-engine-atomic";
import { Provider as StyletronProvider } from "styletron-react";
import { App } from "./App";

const engine = new Styletron();
const { theme } = createTheme(engine, {
  enableDefaultFonts: true,
});

const root = createRoot(document.getElementById("root") || document.body);

root.render(
  <StyletronProvider value={engine}>
    <BaseProvider theme={theme}>
      <App />
    </BaseProvider>
  </StyletronProvider>,
);
