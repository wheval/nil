{ lib
, stdenv
, biome
, callPackage
, npmHooks
, nodejs
, enableTesting ? false
}:

stdenv.mkDerivation rec {
  name = "wallet extension";
  pname = "walletExtension";
  src = lib.sourceByRegex ./.. [
    "package.json"
    "package-lock.json"
    "^niljs(/.*)?$"
    "^smart-contracts(/.*)?$"
    "biome.json"
    "^wallet-extension(/.*)?$"
  ];

  npmDeps = (callPackage ./npmdeps.nix { });

  NODE_PATH = "$npmDeps";

  nativeBuildInputs = [
    nodejs
    npmHooks.npmConfigHook
    biome
  ];

  dontConfigure = true;

  buildPhase = ''
    patchShebangs wallet-extension/node_modules

    (cd smart-contracts; npm run build)
    (cd niljs; npm run build)

    (cd wallet-extension; npm run build)
  '';

  doCheck = enableTesting;

  checkPhase = ''
    export BIOME_BINARY=${biome}/bin/biome

    echo "Checking wallet extension"
    (cd wallet-extension; npm run lint;)

    echo "tests finished successfully"
  '';

  installPhase = ''
    mkdir -p $out
    mv wallet-extension/ $out/extension
  '';
}
