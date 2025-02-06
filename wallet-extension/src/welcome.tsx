import { ErrorPage, createTheme } from "@nilfoundation/ui-kit";
import { BaseProvider } from "baseui";
import { createRoot } from "react-dom/client";
import { ErrorBoundary } from "react-error-boundary";
import { I18nextProvider } from "react-i18next";
import { Client as Styletron } from "styletron-engine-atomic";
import { Provider as StyletronProvider } from "styletron-react";
import { Container } from "./features/components/shared";
import { i18n } from "./i18n.ts";
import { Welcome } from "./pages/welcome";

const engine = new Styletron();
const { theme } = createTheme(engine, {
  enableDefaultFonts: true,
});

const root = createRoot(document.getElementById("root") || document.body);

root.render(
  <StyletronProvider value={engine}>
    <BaseProvider theme={theme}>
      <I18nextProvider i18n={i18n}>
        <Container>
          <ErrorBoundary
            fallback={
              <ErrorPage
                errorDescription="Something went wrong... Please reload the page or try again later."
                errorCode={500}
                redirectPath="https://docs.nil.foundation/"
                redirectTitle="Visit =nil; Documentation"
              />
            }
          >
            <Welcome />
          </ErrorBoundary>
        </Container>
      </I18nextProvider>
    </BaseProvider>
  </StyletronProvider>,
);
