async function loadTutorials() {
  const [testTutorial, testContracts] = await Promise.all([
    import("./assets/testTutorial.md?raw"),
    import("./assets/testContracts.sol?raw"),
  ]);
  const tutorials = [
    {
      stage: 1,
      text: testTutorial.default,
      contracts: testContracts.default,
    },
  ];
  return tutorials;
}

export default loadTutorials;
