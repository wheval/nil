package l1

//go:generate go run github.com/matryer/moq -out eth_client_generated_mock.go -rm -stub -with-resets . EthClient
//go:generate go run github.com/matryer/moq -out l1_contract_generated_mock.go -rm -stub -with-resets . L1Contract
