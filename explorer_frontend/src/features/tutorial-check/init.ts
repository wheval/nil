import { sample } from "effector";
import { $smartAccount } from "../account-connector/model";
import { loadedTutorialPage } from "../code/model";
import { deploySmartContractFx } from "../contracts/models/base";
import { notFoundRoute } from "../routing/routes/routes";
import { tutorialWithUrlStringRoute } from "../routing/routes/tutorialRoute";
import {
  $tutorialCheck,
  deployTutorialContract,
  fetchTutorialCheckEvent,
  fetchTutorialCheckFx,
  runTutorialCheck,
  runTutorialCheckFx,
} from "./model";

$tutorialCheck.on(fetchTutorialCheckFx.doneData, (_, tutorialCheck) => tutorialCheck);

sample({
  clock: runTutorialCheck,
  source: $tutorialCheck,
  target: runTutorialCheckFx,
});

sample({
  clock: [loadedTutorialPage, tutorialWithUrlStringRoute.$params],
  source: tutorialWithUrlStringRoute.$params,
  fn: (params) => params.urlSlug,
  filter: (stage) => stage !== undefined,
  target: fetchTutorialCheckFx,
});

sample({
  clock: fetchTutorialCheckEvent,
  source: fetchTutorialCheckEvent,
  fn: (tutorialCheck) => tutorialCheck.urlSlug,
  target: fetchTutorialCheckFx,
});

sample({
  source: $smartAccount,
  filter: $smartAccount.map((x) => !!x),
  clock: deployTutorialContract,
  fn: ({ smartAccount }, payload) => ({
    app: payload.app,
    args: payload.customArgs as unknown[],
    shardId: payload.customShardId,
    smartAccount: smartAccount!,
  }),
  target: deploySmartContractFx,
});

sample({
  clock: fetchTutorialCheckFx.failData,
  fn: () => notFoundRoute.open(),
});
