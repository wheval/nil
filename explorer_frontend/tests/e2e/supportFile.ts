before(() => {
  cy.log("Starting the test suite...");
});

beforeEach(() => {
  cy.clearCookies();
  cy.clearLocalStorage();
});
