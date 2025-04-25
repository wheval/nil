package contracts

import "embed"

//go:generate bash -c "solc ../../smart-contracts/contracts/*.sol --bin --abi --hashes --overwrite -o ./compiled --no-cbor-metadata --metadata-hash none"
//go:generate bash -c "solc solidity/system/*.sol --bin --abi --hashes --overwrite -o ./compiled/system --allow-paths ./solidity/lib --no-cbor-metadata --metadata-hash none"
//go:generate bash -c "solc solidity/tests/*.sol --allow-paths ../../ --base-path ../../ --bin --abi --hashes --overwrite -o ./compiled/tests --no-cbor-metadata --metadata-hash none"
//go:generate bash -c "ln -nsf ../.. @nilfoundation && solc ../../uniswap/contracts/*.sol --bin --abi --overwrite -o ./compiled/uniswap --allow-paths .,../.. --via-ir && rm @nilfoundation"
//go:embed compiled/*
var Fs embed.FS
