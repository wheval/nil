beforeEach(() => {
  cy.visit("/");
});

describe("search", () => {
  const address = `0x${"a".repeat(40)}`;

  it("search renders", () => {
    cy.get("[placeholder='Search by Address, Transaction Hash, Block Shard ID and Height']").should(
      "exist",
    );
  });

  it("incorrect search", () => {
    cy.get("[placeholder='Search by Address, Transaction Hash, Block Shard ID and Height']").type(
      "0x123",
    );
    cy.get("[data-testid='search-result']").should("exist");
    cy.get("[data-testid='search-result']").should("contain", "No results found");
  });

  it("search for an address", () => {
    cy.get("[placeholder='Search by Address, Transaction Hash, Block Shard ID and Height']").type(
      address,
    );
    cy.get("[data-testid='search-result']").should("exist");
    cy.contains(address).should("exist");
  });

  it("clears search", () => {
    cy.get("[placeholder='Search by Address, Transaction Hash, Block Shard ID and Height']").type(
      address,
    );
    cy.get("[data-testid='search-result']").should("exist");

    cy.get("svg[title='Clear value']").click();
    cy.get("[data-testid='search-result']").should("not.exist");
  });
});
