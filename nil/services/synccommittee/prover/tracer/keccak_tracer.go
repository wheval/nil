package tracer

import (
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/tracing"
	"github.com/NilFoundation/nil/nil/internal/vm"
)

type KeccakBuffer struct {
	buf  []byte
	hash common.Hash
}

type KeccakTracer struct {
	finalizer func()

	hashes []KeccakBuffer
}

func NewKeccakTracer() *KeccakTracer {
	return &KeccakTracer{}
}

func (kt *KeccakTracer) TraceOp(opCode vm.OpCode, opCtx tracing.OpContext) (bool, error) {
	if kt.finalizer != nil {
		return false, ErrTraceNotFinalized
	}

	if opCode != vm.KECCAK256 {
		return false, nil
	}

	stack := NewStackAccessor(opCtx.StackData())
	var (
		memOffset = stack.PopUint64()
		bufSize   = stack.PopUint64()
	)

	buf := getDataOverflowSafe(opCtx.MemoryData(), memOffset, bufSize)
	kt.hashes = append(kt.hashes, KeccakBuffer{
		buf: buf,
	})

	finIdx := len(kt.hashes) - 1
	finStack := NewStackAccessor(opCtx.StackData())
	finStack.Skip(1)
	kt.finalizer = func() {
		trace := &kt.hashes[finIdx]
		trace.hash.SetBytes(finStack.Pop().Bytes())
	}

	return true, nil
}

func (kt *KeccakTracer) FinishPrevOpcodeTracing() {
	if kt.finalizer == nil {
		return
	}

	kt.finalizer()
	kt.finalizer = nil
}

func (kt *KeccakTracer) Finalize() []KeccakBuffer {
	kt.FinishPrevOpcodeTracing()
	return kt.hashes
}
