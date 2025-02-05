package ibft

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/go-ibft/core"
	"github.com/NilFoundation/nil/nil/go-ibft/messages"
	protoIBFT "github.com/NilFoundation/nil/nil/go-ibft/messages/proto"
	"github.com/NilFoundation/nil/nil/internal/config"
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
	Validator  validator
	NetManager *network.Manager
	PrivateKey *ecdsa.PrivateKey
}

type validator interface {
	BuildProposal(ctx context.Context, tx db.RoTx) (*execution.Proposal, error)
	VerifyProposal(ctx context.Context, proposal *execution.Proposal) (*types.Block, error)
	InsertProposal(ctx context.Context, proposal *execution.Proposal, sig types.Signature) error
}

type backendIBFT struct {
	ctx          context.Context
	transportCtx context.Context
	db           db.DB
	consensus    *core.IBFT
	shardId      types.ShardId
	validator    validator
	logger       zerolog.Logger
	nm           *network.Manager
	transport    transport
	signer       *Signer
	mainBlockMap sync.Map
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
	tx, err := i.db.CreateRoTx(i.ctx)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	proposal, err := i.validator.BuildProposal(i.ctx, tx)
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
	i.logger.Debug().
		Uint64(logging.FieldBlockNumber, proposalBlock.PrevBlockId.Uint64()+1).
		Uint64(logging.FieldRound, proposal.Round).
		Uint32(logging.FieldShardId, uint32(i.shardId)).
		Msg("Inserting proposal")

	var signature types.Signature
	for _, seal := range committedSeals {
		if len(seal.Signature) != 0 {
			signature = seal.Signature
		}
	}

	if err := i.validator.InsertProposal(i.ctx, proposalBlock, signature); err != nil {
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
		validator: cfg.Validator,
		logger:    logger,
		nm:        cfg.NetManager,
		signer:    NewSigner(cfg.PrivateKey),
	}
	backend.consensus = core.NewIBFT(l, backend, backend)
	return backend
}

func (i *backendIBFT) Init(ctx context.Context) error {
	i.transportCtx = ctx
	if i.nm == nil {
		i.setupLocalTransport()
		return nil
	}
	return i.setupTransport(ctx)
}

func (i *backendIBFT) GetVotingPowers(height uint64) (map[string]*big.Int, error) {
	// Here we take the latest config, but we should take the config based ot proposer that isn't available at this point
	// TODO(@isergeyam): I think we should rewrite the ibft/core part to get voting powers after the proposer is calculated
	validators, err := i.getValidators(i.ctx, nil)
	if err != nil {
		i.logger.Error().
			Err(err).
			Uint64(logging.FieldHeight, height).
			Msg("Failed to get validators")
		return nil, err
	}

	result := make(map[string]*big.Int, len(validators))
	for _, v := range validators {
		result[string(v.PublicKey[:])] = big.NewInt(1)
	}
	return result, nil
}

func (i *backendIBFT) getValidators(ctx context.Context, mainBlockHash *common.Hash) (validators []config.ValidatorInfo, err error) {
	tx, err := i.db.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	configAccessor, err := config.NewConfigAccessorTx(ctx, tx, mainBlockHash)
	if err != nil {
		return nil, err
	}

	validatorsList, err := config.GetParamValidators(configAccessor)
	if err != nil {
		return nil, err
	}

	return validatorsList.Validators[i.shardId].List, nil
}

func (i *backendIBFT) RunSequence(ctx context.Context, height uint64) error {
	i.ctx = ctx
	i.consensus.RunSequence(ctx, height)
	return nil
}
