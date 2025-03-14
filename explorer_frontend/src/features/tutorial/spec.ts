import { TutorialLevel } from "./model";

async function loadTutorials() {
  const [tutorialOneText, tutorialOneContracts, tutorialOneIcon] = await Promise.all([
    import("./assets/tutorialOne/tutorialOneText.md?raw"),
    import("./assets/tutorialOne/tutorialOneContracts.sol?raw"),
    import("./assets/tutorialOne/tutorialOneIcon.svg"),
  ]);

  const [tutorialTwoText, tutorialTwoContracts, tutorialTwoIcon] = await Promise.all([
    import("./assets/tutorialTwo/tutorialTwoText.md?raw"),
    import("./assets/tutorialTwo/tutorialTwoContracts.sol?raw"),
    import("./assets/tutorialTwo/tutorialTwoIcon.svg"),
  ]);

  const [tutorialThreeText, tutorialThreeContracts, tutorialThreeIcon] = await Promise.all([
    import("./assets/tutorialThree/tutorialThreeText.md?raw"),
    import("./assets/tutorialThree/tutorialThreeContracts.sol?raw"),
    import("./assets/tutorialThree/tutorialThreeIcon.svg"),
  ]);
  const tutorials = [
    {
      stage: 1,
      text: tutorialOneText.default,
      contracts: tutorialOneContracts.default,
      icon: tutorialOneIcon.default,
      completionTime: "5 minutes",
      level: TutorialLevel.Easy,
      title: "Async calls and default tokens",
      description: "Send an async call between shards.",
      urlSlug: "async-call",
    },
    {
      stage: 2,
      text: tutorialTwoText.default,
      contracts: tutorialTwoContracts.default,
      icon: tutorialTwoIcon.default,
      completionTime: "8 minutes",
      level: TutorialLevel.Easy,
      title: "Working with custom tokens",
      description: "Mint and send custom tokens across shards.",
      urlSlug: "custom-tokens",
    },
    {
      stage: 3,
      text: tutorialThreeText.default,
      contracts: tutorialThreeContracts.default,
      icon: tutorialThreeIcon.default,
      completionTime: "7 minutes",
      level: TutorialLevel.Medium,
      title: "Request/response pattern",
      description: "Manage complex async calls and responses to them.",
      urlSlug: "request-response",
    },
  ];
  return tutorials;
}

export default loadTutorials;
