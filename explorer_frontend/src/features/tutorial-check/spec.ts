import runTutorialCheckOne from "./checks/tutorialOneCheck";
import runTutorialCheckThree from "./checks/tutorialThreeCheck";
import runTutorialCheckTwo from "./checks/tutorialTwoCheck";

export const spec = [
  {
    urlSlug: "async-call",
    check: runTutorialCheckOne,
  },
  {
    urlSlug: "custom-tokens",
    check: runTutorialCheckTwo,
  },
  {
    urlSlug: "request-response",
    check: runTutorialCheckThree,
  },
];
