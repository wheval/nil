package mpttracer

import (
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/mpt"
	"github.com/NilFoundation/nil/nil/internal/types"
)

// TracerAccount extends AccountState with additional tracing capabilities
type TracerAccount struct {
	execution.AccountState
	SlotInitialValue execution.Storage
}

func NewTracerAccount(
	es execution.IExecutionState,
	addr types.Address,
	account *types.SmartContract,
) (*TracerAccount, error) {
	as, err := execution.NewAccountState(es, addr, account)
	if err != nil {
		return nil, err
	}

	return &TracerAccount{
		AccountState:     *as,
		SlotInitialValue: make(execution.Storage),
	}, nil
}

// SetState tracks state changes with initial values
func (ta *TracerAccount) SetState(key common.Hash, value common.Hash) error {
	_, exists := ta.SlotInitialValue[key]
	if exists {
		// If an update took place earlier, each subsequent update will be applied.
		// However, the final value might still remain the same as the initial one,
		// which we handle in `GetUpdatedStateSlots`.
		return ta.AccountState.SetState(key, value)
	}
	initialValue, err := ta.AccountState.GetState(key)
	if err != nil {
		return err
	}
	// If the value remains unchanged, skip it and do not
	// add the initial value to the map.
	if value == initialValue {
		return nil
	}

	err = ta.AccountState.SetState(key, value)
	if err != nil {
		return err
	}

	ta.SlotInitialValue[key] = initialValue
	return nil
}

// getUpdatedStateSlots retrieves slots that have been updated
func (ta *TracerAccount) getUpdatedStateSlots() map[common.Hash]MPTValueDiff {
	updatedStateSlots := make(map[common.Hash]MPTValueDiff)

	// Since `SlotInitialValue` is populated inside `SetState`, all updated slots
	// will be stored there. If the slot was touched (both get/set), it will be also stored in `State`.
	for key, initialValue := range ta.SlotInitialValue {
		currentValue := ta.State[key]
		if currentValue == initialValue {
			continue
		}
		updatedStateSlots[key] = MPTValueDiff{
			before: types.Uint256(*initialValue.Uint256()),
			after:  types.Uint256(*currentValue.Uint256()),
		}
	}
	return updatedStateSlots
}

// GetSlotUpdatesTraces generates traces for slot updates
func (ta *TracerAccount) GetSlotUpdatesTraces() ([]StorageTrieUpdateTrace, error) {
	updatedStateSlots := ta.getUpdatedStateSlots()
	traces := make([]StorageTrieUpdateTrace, 0, len(updatedStateSlots))
	for mptKey, diff := range updatedStateSlots {
		// ReadMPTOperation plays no role here, could be any
		proof, err := mpt.BuildProof(ta.StorageTree.Reader, mptKey.Bytes(), mpt.ReadMPTOperation)
		if err != nil {
			return nil, err
		}
		slotChangeTrace := StorageTrieUpdateTrace{
			RootBefore:  ta.StorageTree.RootHash(),
			ValueBefore: diff.before,
			PathBefore:  proof.PathToNode,
			Proof:       proof,
		}

		if err := ta.StorageTree.Update(mptKey, &diff.after); err != nil {
			return nil, err
		}

		proof, err = mpt.BuildProof(ta.StorageTree.Reader, mptKey.Bytes(), mpt.ReadMPTOperation)
		if err != nil {
			return nil, err
		}
		slotChangeTrace.PathAfter = proof.PathToNode
		slotChangeTrace.RootAfter = ta.StorageTree.RootHash()
		slotChangeTrace.ValueAfter = diff.after

		traces = append(traces, slotChangeTrace)
	}
	return traces, nil
}
