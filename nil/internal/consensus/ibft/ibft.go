package ibft

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/go-ibft/core"
	"github.com/NilFoundation/nil/nil/go-ibft/messages"
	protoIBFT "github.com/NilFoundation/nil/nil/go-ibft/messages/proto"
	"github.com/NilFoundation/nil/nil/internal/collate"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/rs/zerolog"
)

const ibftProto = "/ibft/0.2"

type ConsensusParams struct {
	ShardId    types.ShardId
	Db         db.DB
	Scheduler  *collate.Scheduler
	NetManager *network.Manager
	PrivateKey *ecdsa.PrivateKey
}

type backendIBFT struct {
	ctx       context.Context
	db        db.DB
	consensus *core.IBFT
	shardId   types.ShardId
	scheduler *collate.Scheduler
	logger    zerolog.Logger
	nm        *network.Manager
	transport transport
	signer    *Signer
}

var _ core.Backend = &backendIBFT{}

func (i *backendIBFT) unmarshalProposal(raw []byte) (*execution.Proposal, error) {
	proposal := &execution.Proposal{}
	if err := proposal.UnmarshalSSZ(raw); err != nil {
		return nil, err
	}
	return proposal, nil
}

func (i *backendIBFT) BuildProposal(view *protoIBFT.View) []byte {
	proposal, err := i.scheduler.BuildProposal(i.ctx)
	if err != nil {
		return nil
	}
	data, err := proposal.MarshalSSZ()
	if err != nil {
		return nil
	}
	return data
}

func (i *backendIBFT) InsertProposal(proposal *protoIBFT.Proposal, committedSeals []*messages.CommittedSeal) {
	proposalBlock, err := i.unmarshalProposal(proposal.RawProposal)
	if err != nil {
		return
	}

	var signature types.Signature
	for _, seal := range committedSeals {
		if len(seal.Signature) != 0 {
			signature = seal.Signature
		}
	}

	if err := i.scheduler.InsertProposal(i.ctx, proposalBlock, signature); err != nil {
		i.logger.Error().Err(err).Msg("fail to insert proposal")
	}
}

func (i *backendIBFT) ID() []byte {
	return i.signer.GetPublicKey()
}

func (i *backendIBFT) isActiveValidator() bool {
	return true
}

func NewConsensus(cfg *ConsensusParams) *backendIBFT {
	logger := logging.NewLogger("consensus").With().Stringer(logging.FieldShardId, cfg.ShardId).Logger()
	l := &ibftLogger{
		logger: logger.With().CallerWithSkipFrameCount(3).Logger(),
	}

	backend := &backendIBFT{
		db:        cfg.Db,
		shardId:   cfg.ShardId,
		scheduler: cfg.Scheduler,
		logger:    logger,
		nm:        cfg.NetManager,
		signer:    NewSigner(cfg.PrivateKey),
	}
	backend.consensus = core.NewIBFT(l, backend, backend)
	return backend
}

func (i *backendIBFT) GetVotingPowers(height uint64) (map[string]*big.Int, error) {
	result := make(map[string]*big.Int)
	result[string(i.ID())] = big.NewInt(1)
	return result, nil
}

func (i *backendIBFT) Init(ctx context.Context) error {
	if i.nm == nil {
		i.setupLocalTransport()
		return nil
	}
	return i.setupTransport(ctx)
}

func (i *backendIBFT) RunSequence(ctx context.Context, height uint64) error {
	i.ctx = ctx
	i.consensus.RunSequence(ctx, height)
	return nil
}
