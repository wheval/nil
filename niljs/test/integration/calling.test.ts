import { waitTillCompleted } from "../../src/index.js";
import { generateRandomAddress, generateTestSmartAccount, newClient } from "./helpers.js";

const client = newClient();

test("Async call to another shard send value", async () => {
  const smartAccount = await generateTestSmartAccount(1);

  const anotherAddress = generateRandomAddress(2);

  const gasPriceOnShard2 = await client.getGasPrice(2);

  const hash = await smartAccount.sendTransaction({
    to: anotherAddress,
    value: 50_000_000n,
    feeCredit: 100_000n * gasPriceOnShard2,
  });

  const receipts = await waitTillCompleted(client, hash);

  expect(receipts).toBeDefined();
  expect(receipts.some((r) => !r.success)).toBe(false);

  const balance = await client.getBalance(anotherAddress, "latest");
  expect(balance).toBeGreaterThan(0n);
});

test("sync call same shard send value", async () => {
  const smartAccount = await generateTestSmartAccount(1);

  const anotherAddress = generateRandomAddress(1);

  const hash = await smartAccount.syncSendTransaction({
    to: anotherAddress,
    value: 10n,
    gas: 100000n,
  });

  const receipts = await waitTillCompleted(client, hash);

  expect(receipts).toBeDefined();
  expect(receipts.some((r) => !r.success)).toBe(false);
  const balance = await client.getBalance(anotherAddress, "latest");

  expect(balance).toBe(10n);
});
