import { FaucetClient, HttpTransport } from "@nilfoundation/niljs";
import { sample } from "effector";
import { $rpcUrl } from "../account-connector/model";
import { $faucets, fetchFaucetsEvent, fetchFaucetsFx } from "./model";

fetchFaucetsFx.use(async (rpcUrl) => {
  const faucetClient = new FaucetClient({
    transport: new HttpTransport({ endpoint: rpcUrl }),
  });

  return await faucetClient.getAllFaucets();
});

sample({
  clock: fetchFaucetsEvent,
  source: $rpcUrl,
  target: fetchFaucetsFx,
});

$rpcUrl.watch((rpcUrl) => {
  if (rpcUrl) {
    fetchFaucetsEvent();
  }
});

$faucets.on(fetchFaucetsFx.doneData, (_, balance) => balance);

fetchFaucetsEvent();
