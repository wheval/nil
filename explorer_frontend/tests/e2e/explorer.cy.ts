beforeEach(() => {
  cy.visit("/");
});

describe("home page", () => {
  it("the h3 contains the correct text", () => {
    cy.get("h3").should("contain", "Secure Ethereum scaling");
  });
});

describe("renders api responses correctly", () => {
  it("blocks table", () => {
    cy.intercept("GET", "https://explore.nil.foundation/api/block.latestBlocks", {
      fixture: "latestBlocks.json",
    }).as("blocks");

    cy.wait("@blocks", { timeout: 10000 });

    cy.get('[data-testid="blocks-table"]').within(() => {
      cy.get("tbody").children().should("exist");
      cy.get("tbody").children().should("have.length", 3);
    });
  });

  it("shards stat", () => {
    cy.intercept("GET", "https://explore.nil.foundation/api/shards.shardsStat", {
      fixture: "shardStat.json",
    }).as("shards");

    cy.wait("@shards", { timeout: 10000 });

    cy.get("[data-testid='shards-container']").as("shardsContainer").should("exist");
    cy.get("@shardsContainer").children().should("have.length", 5);
  });

  it("transaction chart", () => {
    cy.intercept(
      "GET",
      "https://explore.nil.foundation/api/info.transactionStat?input=%7B%22period%22%3A%2230m%22%7D",
      {
        fixture: "transactions.json",
      },
    ).as("trx");

    cy.wait("@trx", { timeout: 10000 });

    cy.get("[data-testid='transaction-chart']").as("txChart").should("exist");
    cy.get("@txChart").should("exist");

    cy.get("[data-testid='transaction-chart'] canvas").should("exist").should("have.attr", "width");
  });

  it("empty transaction chart", () => {
    cy.intercept(
      "GET",
      "https://explore.nil.foundation/api/info.transactionStat?input=%7B%22period%22%3A%2230m%22%7D",
      {
        result: {
          data: [],
        },
      },
    ).as("trx");

    cy.wait("@trx", { timeout: 10000 });

    cy.get("[data-testid='transaction-chart']").as("txChart").should("exist");
    cy.get("@txChart").should("exist");

    cy.get("[data-testid='transaction-chart'] canvas").should("not.exist");
    cy.get("h4").should("contain", "No data to display");
  });
});
