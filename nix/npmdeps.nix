{ lib, stdenv, fetchNpmDeps }:
let
  inherit (lib) fileset;
in
(fetchNpmDeps {
  src = fileset.toSource {
    root = ./..;
    fileset = fileset.unions [
      ../package-lock.json
      ../package.json
      ../clijs/package.json
      ../docs/package.json
      ../niljs/package.json
      ../create-nil-hardhat-project/package.json
      ../smart-contracts/package.json
      ../explorer_backend/package.json
      ../explorer_frontend/package.json
      ../uniswap/package.json
    ];
  };
  hash = "sha256-Est/QV5+OFPzD82/xdYyZAEmmN5DmhVow54ZE2lSU9A=";
})
