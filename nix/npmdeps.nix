{ lib, stdenv, pnpm_10, nodejs }:
let
  inherit (lib) fileset;
  pnpm = pnpm_10;
in
(pnpm.fetchDeps {
  src = fileset.toSource {
    root = ./..;
    fileset = fileset.unions [
      ../pnpm-workspace.yaml
      ../pnpm-lock.yaml
      ../.npmrc
      ../package.json
      ../clijs/package.json
      ../docs/package.json
      ../niljs/package.json
      ../hardhat-plugin/package.json
      ../create-nil-hardhat-project/package.json
      ../smart-contracts/package.json
      ../explorer_backend/package.json
      ../explorer_frontend/package.json
      ../uniswap/package.json
      ../rollup-bridge-contracts/package.json
      ../wallet-extension/package.json
      ../docs_ai_backend/package.json
    ];
  };
  pname = "nil";
  hash = "sha256-1uCdRQ+1mOu20z171q9p34aYDEHmm0/rVnMoIDiz6dc=";
})
