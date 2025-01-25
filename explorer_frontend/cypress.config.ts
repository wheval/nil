import { defineConfig } from "cypress";

// biome-ignore lint/style/noDefaultExport: <explanation>
export default defineConfig({
  e2e: {
    baseUrl: "http://localhost:3000",
    specPattern: "tests/**/*.cy.ts",
    supportFile: "tests/e2e/supportFile.ts",
    fixturesFolder: "tests/e2e/fixtures",
    screenshotsFolder: "tests/e2e/screenshots",
    downloadsFolder: "tests/e2e/downloads",
  },
});
