import { waitTillCompleted } from "../../src/index.js";
import { generateRandomAddress, generateTestSmartAccount, newClient } from "./helpers.js";

const client = newClient();

test("mint and transfer tokens", async () => {
  const smartAccount = await generateTestSmartAccount();
  const smartAccountAddress = smartAccount.address;

  const mintCount = 100_000_000n;

  {
    const hashTransaction = await smartAccount.setTokenName("MY_TOKEN");
    await waitTillCompleted(client, hashTransaction);
  }

  {
    const hashTransaction = await smartAccount.mintToken(mintCount);
    await waitTillCompleted(client, hashTransaction);
  }

  const tokens = await client.getTokens(smartAccountAddress, "latest");

  expect(tokens).toBeDefined();
  expect(Object.keys(tokens).length).toBeGreaterThan(0);
  expect(tokens[smartAccountAddress]).toBeDefined();
  expect(tokens[smartAccountAddress]).toBe(mintCount);

  const anotherAddress = generateRandomAddress(2);

  const transferCount = 100_000n;

  const gasPriceOnShard2 = await client.getGasPrice(2);
  const sendHash = await smartAccount.sendTransaction({
    to: anotherAddress,
    value: 10_000_000n,
    feeCredit: 100_000n * gasPriceOnShard2,
    tokens: [
      {
        id: smartAccountAddress,
        amount: transferCount,
      },
    ],
  });

  await waitTillCompleted(client, sendHash);

  const anotherTokens = await client.getTokens(anotherAddress, "latest");

  expect(anotherTokens).toBeDefined();
  expect(Object.keys(anotherTokens).length).toBeGreaterThan(0);
  expect(anotherTokens[smartAccountAddress]).toBeDefined();
  expect(anotherTokens[smartAccountAddress]).toBe(transferCount);
});
