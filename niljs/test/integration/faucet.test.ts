import { generateTestSmartAccount, newClient, newFaucetClient, topUpTest } from "./helpers.js";

const client = newClient();

const faucetClient = newFaucetClient();

test("getAllFaucets", async () => {
  const faucets = await faucetClient.getAllFaucets();

  expect(Object.keys(faucets).length).toBeGreaterThan(0);
  expect(faucets.NIL).toBeDefined();
  expect(faucets.BTC).toBeDefined();
});

test("mint tokens", async () => {
  const smartAccount = await generateTestSmartAccount();

  await topUpTest(smartAccount.address, "BTC");

  const tokens = await client.getTokens(smartAccount.address, "latest");

  expect(tokens).toBeDefined();
  expect(Object.keys(tokens).length).toBeGreaterThan(0);

  const faucets = await faucetClient.getAllFaucets();
  expect(tokens[faucets.BTC]).toBeDefined();
});
