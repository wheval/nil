.PHONY: ssz_sszx
ssz_sszx: nil/common/sszx/map_encoding.go

nil/common/sszx/map_encoding.go: nil/common/sszx/map.go
	cd nil/common/sszx && go run github.com/NilFoundation/fastssz/sszgen --path map.go \
		--objs KeyValue,MapHolder
