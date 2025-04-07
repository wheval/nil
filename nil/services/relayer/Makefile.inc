root_services = nil/services
root_relayer = nil/services/relayer

.PHONY: $(root_relayer)/embed_l2_abi
$(root_relayer)/embed_l2_abi:
	@cat $(root_services)/rollup-bridge-contracts-compiled/contracts/bridge/l2/L2BridgeMessenger.sol/L2BridgeMessenger.json | jq .abi > $(root_relayer)/internal/l2/L2BridgeMessenger.json.abi

.PHONY: $(root_relayer)/embed_l1_abi
$(root_relayer)/embed_l1_abi:
	@cat $(root_services)/rollup-bridge-contracts-compiled/contracts/bridge/l1/L1BridgeMessenger.sol/L1BridgeMessenger.json | jq .abi > $(root_relayer)/internal/l1/L1BridgeMessenger.json.abi

.PHONY: $(root_relayer)/generate_l1_abi
$(root_relayer)/generate_l1_abi: $(root_relayer)/embed_l1_abi $(root_relayer)/internal/l1/abi.go
	cd $(root_relayer)/internal/l1 && go generate abi.go

.PHONY: $(root_relayer)/gen_l1_mocks
$(root_relayer)/gen_l1_mocks: $(root_relayer)/generate_l1_abi $(root_relayer)/internal/l1/generate.go
	cd $(root_relayer)/internal/l1 && go generate generate.go

.PHONY: $(root_relayer)/gen_l2_mocks
$(root_relayer)/gen_l2_mocks: $(root_relayer)/embed_l2_abi $(root_relayer)/internal/l2/generate.go
	cd $(root_relayer)/internal/l2 && go generate generate.go
