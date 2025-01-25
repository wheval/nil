import { FaucetClient, HttpTransport } from "@nilfoundation/niljs";
import { sample } from "effector";
import { $endpoint } from "../account-connector/model";
import { $faucets, fetchFaucetsEvent, fetchFaucetsFx } from "./model";

fetchFaucetsFx.use(async (endpoint) => {
  const faucetClient = new FaucetClient({
    transport: new HttpTransport({ endpoint }),
  });

  return await faucetClient.getAllFaucets();
});

sample({
  clock: fetchFaucetsEvent,
  source: $endpoint,
  target: fetchFaucetsFx,
});

$endpoint.watch((endpoint) => {
  if (endpoint) {
    fetchFaucetsEvent();
  }
});

$faucets.on(fetchFaucetsFx.doneData, (_, balance) => balance);

fetchFaucetsEvent();
