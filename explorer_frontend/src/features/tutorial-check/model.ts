import { createDomain } from "effector";
import type { App } from "../../types";
import { spec } from "./spec";

export type TutorialCheck = {
  urlSlug: string;
  check: () => Promise<void>;
};

export const tutorialCheckDomain = createDomain("tutorial-check");

export const $tutorialCheck = tutorialCheckDomain.createStore<TutorialCheck>({
  urlSlug: "",
  check: async () => {},
});

export const deployTutorialContract = tutorialCheckDomain.createEvent<{
  app: App;
  customArgs: Record<string, string | boolean>[];
  customShardId: number;
}>();

export const tutorialContractStepPassedEvent = tutorialCheckDomain.createEvent<string>();

export const tutorialContractStepFailedEvent = tutorialCheckDomain.createEvent<string>();

export const fetchTutorialCheckEvent = tutorialCheckDomain.createEvent<TutorialCheck>();

export const fetchTutorialCheckFx = tutorialCheckDomain.createEffect<string, TutorialCheck>();

export const runTutorialCheck = tutorialCheckDomain.createEvent();

export const runTutorialCheckFx = tutorialCheckDomain.createEffect(
  async (tutorialCheck: TutorialCheck) => {
    return await tutorialCheck.check();
  },
);

fetchTutorialCheckFx.use(async (urlSlug) => {
  const tutorialCheck = spec.find((check) => check.urlSlug === urlSlug);
  if (!tutorialCheck) {
    throw new Error("Tutorial check not found");
  }
  return tutorialCheck;
});
