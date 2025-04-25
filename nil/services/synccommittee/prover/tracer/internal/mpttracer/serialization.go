package mpttracer

import (
	"encoding/hex"
	"errors"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/internal/mpt"
	"github.com/NilFoundation/nil/nil/internal/types"
	pb "github.com/NilFoundation/nil/nil/services/synccommittee/prover/proto"
	"github.com/NilFoundation/nil/nil/services/synccommittee/prover/tracer/internal/constants"
	"github.com/ethereum/go-ethereum/crypto"
)

func TracesFromProto(pbMptTraces *pb.MPTTraces) (*MPTTraces, error) {
	check.PanicIfNot(pbMptTraces != nil)

	addrToStorageTraces := make(map[types.Address][]StorageTrieUpdateTrace,
		len(pbMptTraces.GetStorageTracesByAccount()))
	for addr, pbStorageTrace := range pbMptTraces.GetStorageTracesByAccount() {
		storageTraces := make([]StorageTrieUpdateTrace, len(pbStorageTrace.GetUpdatesTraces()))
		for i, pbStorageUpdateTrace := range pbStorageTrace.GetUpdatesTraces() {
			proof, err := protoToGoProof(pbStorageUpdateTrace.GetSszProof())
			if err != nil {
				return nil, err
			}
			storageTraces[i] = StorageTrieUpdateTrace{
				Key:         common.HexToHash(pbStorageUpdateTrace.GetKey()),
				RootBefore:  common.HexToHash(pbStorageUpdateTrace.GetRootBefore()),
				RootAfter:   common.HexToHash(pbStorageUpdateTrace.GetRootAfter()),
				ValueBefore: pb.ProtoUint256ToUint256(pbStorageUpdateTrace.GetValueBefore()),
				ValueAfter:  pb.ProtoUint256ToUint256(pbStorageUpdateTrace.GetValueAfter()),
				Proof:       proof,
				// PathBefore: // Unused
				// PathAfter:  // Unused
			}
		}
		addrToStorageTraces[types.HexToAddress(addr)] = storageTraces
	}

	contractTrieTraces := make([]ContractTrieUpdateTrace, len(pbMptTraces.GetContractTrieTraces()))
	for i, pbContractTrieUpdate := range pbMptTraces.GetContractTrieTraces() {
		proof, err := protoToGoProof(pbContractTrieUpdate.GetSszProof())
		if err != nil {
			return nil, err
		}
		contractTrieTraces[i] = ContractTrieUpdateTrace{
			Key:         common.HexToHash(pbContractTrieUpdate.GetKey()),
			RootBefore:  common.HexToHash(pbContractTrieUpdate.GetRootBefore()),
			RootAfter:   common.HexToHash(pbContractTrieUpdate.GetRootAfter()),
			ValueBefore: smartContractFromProto(pbContractTrieUpdate.GetValueBefore()),
			ValueAfter:  smartContractFromProto(pbContractTrieUpdate.GetValueAfter()),
			Proof:       proof,
			// PathBefore: // Unused
			// PathAfter:  // Unused
		}
	}

	ret := &MPTTraces{
		StorageTracesByAccount: addrToStorageTraces,
		ContractTrieTraces:     contractTrieTraces,
	}

	return ret, nil
}

func TracesToProto(mptTraces *MPTTraces, traceIdx uint64) (*pb.MPTTraces, error) {
	check.PanicIfNot(mptTraces != nil)

	pbAddrToStorageTraces := make(map[string]*pb.StorageTrieUpdatesTraces, len(mptTraces.StorageTracesByAccount))
	for addr, storageTraces := range mptTraces.StorageTracesByAccount {
		pbStorageTraces := make([]*pb.StorageTrieUpdateTrace, len(storageTraces))
		for i, storageUpdateTrace := range storageTraces {
			proof, err := goToProtoProof(storageUpdateTrace.Proof)
			if err != nil {
				return nil, err
			}
			proofPathBefore, err := proofPathToProto(storageUpdateTrace.PathBefore)
			if err != nil {
				return nil, err
			}
			proofPathAfter, err := proofPathToProto(storageUpdateTrace.PathAfter)
			if err != nil {
				return nil, err
			}

			pbStorageTraces[i] = &pb.StorageTrieUpdateTrace{
				Key:         storageUpdateTrace.Key.Hex(),
				RootBefore:  storageUpdateTrace.RootBefore.Hex(),
				RootAfter:   storageUpdateTrace.RootAfter.Hex(),
				ValueBefore: pb.Uint256ToProtoUint256(storageUpdateTrace.ValueBefore),
				ValueAfter:  pb.Uint256ToProtoUint256(storageUpdateTrace.ValueAfter),
				SszProof:    proof,
				ProofBefore: proofPathBefore,
				ProofAfter:  proofPathAfter,
			}
		}

		pbAddrToStorageTraces[addr.Hex()] = &pb.StorageTrieUpdatesTraces{UpdatesTraces: pbStorageTraces}
	}

	pbContractTrieTraces := make([]*pb.ContractTrieUpdateTrace, len(mptTraces.ContractTrieTraces))
	for i, contractTrieUpdate := range mptTraces.ContractTrieTraces {
		proof, err := goToProtoProof(contractTrieUpdate.Proof)
		if err != nil {
			return nil, err
		}

		proofPathBefore, err := proofPathToProto(contractTrieUpdate.PathBefore)
		if err != nil {
			return nil, err
		}
		proofPathAfter, err := proofPathToProto(contractTrieUpdate.PathAfter)
		if err != nil {
			return nil, err
		}

		pbContractTrieTraces[i] = &pb.ContractTrieUpdateTrace{
			Key:         contractTrieUpdate.Key.Hex(),
			RootBefore:  contractTrieUpdate.RootBefore.Hex(),
			RootAfter:   contractTrieUpdate.RootAfter.Hex(),
			ValueBefore: smartContractToProto(contractTrieUpdate.ValueBefore),
			ValueAfter:  smartContractToProto(contractTrieUpdate.ValueAfter),
			SszProof:    proof,
			ProofBefore: proofPathBefore,
			ProofAfter:  proofPathAfter,
		}
	}

	ret := &pb.MPTTraces{
		StorageTracesByAccount: pbAddrToStorageTraces,
		ContractTrieTraces:     pbContractTrieTraces,
		TraceIdx:               traceIdx,
		ProtoHash:              constants.ProtoHash,
	}

	return ret, nil
}

func goToProtoProof(p mpt.Proof) ([]byte, error) {
	encodedProof, err := p.Encode()
	if err != nil {
		return nil, err
	}
	return encodedProof, nil
}

func protoToGoProof(pbProof []byte) (mpt.Proof, error) {
	return mpt.DecodeProof(pbProof)
}

// smartContractFromProto converts a Protocol Buffers SmartContract to Go SmartContract
func smartContractFromProto(pbSmartContract *pb.SmartContract) *types.SmartContract {
	if pbSmartContract == nil {
		return nil
	}

	var balance types.Value
	if pbSmartContract.GetBalance() != nil {
		b := pb.ProtoUint256ToUint256(pbSmartContract.GetBalance())
		balance = types.NewValue(b.Int())
	}
	return &types.SmartContract{
		Address:          types.HexToAddress(pbSmartContract.GetAddress()),
		Balance:          types.NewValue(balance.Int()),
		TokenRoot:        common.HexToHash(pbSmartContract.GetTokenRoot()),
		StorageRoot:      common.HexToHash(pbSmartContract.GetStorageRoot()),
		CodeHash:         common.HexToHash(pbSmartContract.GetCodeHash()),
		AsyncContextRoot: common.HexToHash(pbSmartContract.GetAsyncContextRoot()),
		Seqno:            types.Seqno(pbSmartContract.GetSeqno()),
		ExtSeqno:         types.Seqno(pbSmartContract.GetExtSeqno()),
	}
}

// smartContractToProto converts a Go SmartContract to Protocol Buffers SmartContract
func smartContractToProto(smartContract *types.SmartContract) *pb.SmartContract {
	if smartContract == nil {
		return nil
	}

	var pbBalance *pb.Uint256
	if smartContract.Balance.Uint256 != nil {
		pbBalance = pb.Uint256ToProtoUint256(smartContract.Balance.Uint256)
	}
	return &pb.SmartContract{
		Address:          smartContract.Address.Hex(),
		Balance:          pbBalance,
		TokenRoot:        smartContract.TokenRoot.Hex(),
		StorageRoot:      smartContract.StorageRoot.Hex(),
		CodeHash:         smartContract.CodeHash.Hex(),
		AsyncContextRoot: smartContract.AsyncContextRoot.Hex(),
		Seqno:            uint64(smartContract.Seqno),
		ExtSeqno:         uint64(smartContract.ExtSeqno),
	}
}

func proofPathToProto(pathToNode mpt.SimpleProof) (*pb.HumanReadableProof, error) {
	nodes := make([]*pb.Node, len(pathToNode))
	for i, node := range pathToNode {
		pbNode, err := nodeToProto(node)
		if err != nil {
			return nil, err
		}
		nodes[i] = pbNode
	}
	return &pb.HumanReadableProof{Nodes: nodes}, nil
}

func nodeToProto(node mpt.Node) (*pb.Node, error) {
	switch n := node.(type) {
	case *mpt.LeafNode:
		return leafNodeToProto(n)
	case *mpt.ExtensionNode:
		return extensionNodeToProto(n)
	case *mpt.BranchNode:
		return branchNodeToProto(n)
	default:
		return nil, errors.New("unknown node type")
	}
}

func leafNodeToProto(node *mpt.LeafNode) (*pb.Node, error) {
	hash, err := nodeHash(node)
	if err != nil {
		return nil, err
	}
	return &pb.Node{
		NodeType: &pb.Node_Leaf{
			Leaf: &pb.LeafNode{
				Key:   node.NodePath.Hex(),
				Value: hex.EncodeToString(node.LeafData),
			},
		},
		Hash: hash.Hex(),
	}, nil
}

func extensionNodeToProto(node *mpt.ExtensionNode) (*pb.Node, error) {
	hash, err := nodeHash(node)
	if err != nil {
		return nil, err
	}
	return &pb.Node{
		NodeType: &pb.Node_Extension{
			Extension: &pb.ExtensionNode{
				Prefix:   node.NodePath.Hex(),
				NextHash: common.BytesToHash(node.NextRef).Hex(),
			},
		},
		Hash: hash.Hex(),
	}, nil
}

func referencesToHex(references []mpt.Reference) []string {
	protoBytes := make([]string, len(references))
	for i, ref := range references {
		protoBytes[i] = common.BytesToHash(ref).Hex() // since Reference is []byte, this assignment is direct
	}
	return protoBytes
}

func branchNodeToProto(node *mpt.BranchNode) (*pb.Node, error) {
	hash, err := nodeHash(node)
	if err != nil {
		return nil, err
	}
	return &pb.Node{
		NodeType: &pb.Node_Branch{
			Branch: &pb.BranchNode{
				ChildHashes: referencesToHex(node.Branches[:]),
				Value:       hex.EncodeToString(node.Value),
			},
		},
		Hash: hash.Hex(),
	}, nil
}

// Helper function to compute the hash of node
func nodeHash(node mpt.Node) (common.Hash, error) {
	data, err := node.Encode()
	if err != nil {
		return common.EmptyHash, err
	}

	if len(data) < 32 {
		return common.EmptyHash, nil
	}

	return common.BytesToHash(crypto.Keccak256(data)), nil
}
