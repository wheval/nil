package types

//go:generate go run github.com/NilFoundation/fastssz/sszgen --path log.go -include ../../common/hexutil/bytes.go,../../common/length.go,address.go,../../common/hash.go,block.go,uint256.go --objs Log,DebugLog
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path receipt.go -include ../../common/hexutil/bytes.go,../../common/length.go,address.go,gas.go,value.go,block.go,bloom.go,log.go,transaction.go,exec_errors.go,../../common/hash.go,uint256.go --objs Receipt
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path transaction.go -include ../../common/length.go,address.go,gas.go,value.go,code.go,shard.go,bloom.go,log.go,../../common/hash.go,signature.go,account.go,bitflags.go --objs Transaction,ExternalTransaction,InternalTransactionPayload,TransactionDigest,TransactionFlags,EvmState,AsyncContext,AsyncResponsePayload
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path block.go -include ../../common/length.go,signature.go,address.go,code.go,shard.go,bloom.go,log.go,value.go,transaction.go,../../common/hash.go --objs BlockData,Block
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path collator.go -include shard.go,block.go,transaction.go --objs Neighbor,CollatorState
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path account.go -include ../../common/length.go,transaction.go,address.go,value.go,code.go,shard.go,bloom.go,log.go,../../common/hash.go --objs SmartContract,TokenBalance
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path version_info.go -include ../../common/hash.go,../../common/length.go --objs VersionInfo
//go:generate stringer -type=ErrorCode -trimprefix=Error
