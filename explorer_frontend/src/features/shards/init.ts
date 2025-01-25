import { sample } from "effector";
import { interval } from "patronum";
import { explorerRoute } from "../routing/routes/explorerRoute";
import { $shards, fetchShardsFx } from "./models/model";

const { tick } = interval({
  timeout: 1000 * 6,
  start: explorerRoute.navigated,
  stop: explorerRoute.closed,
  leading: true,
});

sample({
  clock: tick,
  target: fetchShardsFx,
});

$shards.reset(explorerRoute.navigated);

$shards.on(fetchShardsFx.doneData, (_, data) => data);
