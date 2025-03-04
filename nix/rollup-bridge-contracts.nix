{ lib
, stdenv
, biome
, callPackage
, npmHooks
, nodejs
, nil
, dotenv-cli
, enableTesting ? false
}:

stdenv.mkDerivation rec {
  name = "rollup-bridge-contracts";
  pname = "rollup-bridge-contracts";
  src = lib.sourceByRegex ./.. [
    "package.json"
    "package-lock.json"
    "^niljs(/.*)?$"
    "^rollup-bridge-contracts(/.*)?$"
    "biome.json"
    "^create-nil-hardhat-project(/.*)?$"
  ];

  npmDeps = (callPackage ./npmdeps.nix { });

  nativeBuildInputs = [
    nodejs
    npmHooks.npmConfigHook
  ];

  soljson26 = builtins.fetchurl {
    url = "https://binaries.soliditylang.org/wasm/soljson-v0.8.26+commit.8a97fa7a.js";
    sha256 = "1mhww44ni55yfcyn4hjql2hwnvag40p78kac7jjw2g2jdwwyb1fv";
  };

  buildPhase = ''
    echo "Installing soljson"
    (cd create-nil-hardhat-project; bash install_soljson.sh ${soljson26})
    export BIOME_BINARY=${biome}/bin/biome

    cd rollup-bridge-contracts
    pwd
    cp .env.example .env

    export GETH_PRIVATE_KEY=002f28996b406c557ff579766af59ba66a3f103b8b90de6e9baad8ae211c0071
    export GETH_WALLET_ADDRESS=0xc8d5559BA22d11B0845215a781ff4bF3CCa0EF89

    npx dotenv -e .env -- npx replace-in-file 'GETH_PRIVATE_KEY=""' "GETH_PRIVATE_KEY=$GETH_PRIVATE_KEY" .env
    npx dotenv -e .env -- npx replace-in-file 'GETH_WALLET_ADDRESS=""' "GETH_WALLET_ADDRESS=$GETH_WALLET_ADDRESS" .env
    echo "start compiling"
    npx hardhat clean && npx hardhat compile
  '';

  installPhase = ''
    mkdir -p $out
    cp -r * $out/
  '';
}
