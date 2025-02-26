package collate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/NilFoundation/nil/nil/common/assert"
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/rpc/rawapi/pb"
	"github.com/multiformats/go-multistream"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

type SyncerConfig struct {
	execution.BlockGeneratorParams

	Name            string
	ShardId         types.ShardId
	Timeout         time.Duration // pull blocks if no new blocks appear in the topic for this duration
	BootstrapPeers  []network.AddrInfo
	ZeroStateConfig *execution.ZeroStateConfig
}

// every n-th block will be reported to info log (to avoid spamming)
const blockReportInterval = 100

type Syncer struct {
	config *SyncerConfig

	topic string

	db             db.DB
	networkManager *network.Manager

	logger zerolog.Logger

	waitForSync *sync.WaitGroup

	subsMutex sync.Mutex
	subsId    uint64
	subs      map[uint64]chan types.BlockNumber
	validator *Validator
}

func NewSyncer(cfg *SyncerConfig, validator *Validator, db db.DB, networkManager *network.Manager) (*Syncer, error) {
	var waitForSync sync.WaitGroup
	waitForSync.Add(1)

	return &Syncer{
		config:         cfg,
		topic:          topicShardBlocks(cfg.ShardId),
		db:             db,
		networkManager: networkManager,
		logger: logging.NewLogger(cfg.Name).With().
			Stringer(logging.FieldShardId, cfg.ShardId).
			Logger(),
		waitForSync: &waitForSync,
		subs:        make(map[uint64]chan types.BlockNumber),
		validator:   validator,
	}, nil
}

func (s *Syncer) shardIsEmpty(ctx context.Context) (bool, error) {
	block, _, err := s.validator.GetLastBlock(ctx)
	if err != nil {
		return false, err
	}
	return block == nil, nil
}

func (s *Syncer) WaitComplete() {
	s.waitForSync.Wait()
}

func (s *Syncer) Subscribe() (uint64, <-chan types.BlockNumber) {
	s.subsMutex.Lock()
	defer s.subsMutex.Unlock()

	ch := make(chan types.BlockNumber, 1)
	id := s.subsId
	s.subs[id] = ch
	s.subsId++
	return id, ch
}

func (s *Syncer) Unsubscribe(id uint64) {
	s.subsMutex.Lock()
	defer s.subsMutex.Unlock()

	close(s.subs[id])
	delete(s.subs, id)
}

func (s *Syncer) notify(blockId types.BlockNumber) {
	s.subsMutex.Lock()
	defer s.subsMutex.Unlock()

	for _, ch := range s.subs {
		ch <- blockId
	}
}

func (s *Syncer) FetchSnapshot(ctx context.Context) error {
	if snapIsRequired, err := s.shardIsEmpty(ctx); err != nil {
		return err
	} else if snapIsRequired {
		for _, peer := range s.config.BootstrapPeers {
			if err = fetchSnapshot(ctx, s.networkManager, &peer, s.db, s.logger); err == nil {
				return nil
			}
		}
	}
	return nil
}

func (s *Syncer) SetBootstrapHandler(ctx context.Context) {
	// Enable handler for snapshot relaying
	SetBootstrapHandler(ctx, s.networkManager, s.db)
}

func (s *Syncer) Run(ctx context.Context) error {
	if s.networkManager == nil {
		s.waitForSync.Done()
		return nil
	}

	block, hash, err := s.validator.GetLastBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to read last block number: %w", err)
	}

	s.logger.Debug().
		Stringer(logging.FieldBlockHash, hash).
		Uint64(logging.FieldBlockNumber, uint64(block.Id)).
		Msg("Initialized syncer at starting block")

	s.logger.Info().Msg("Starting sync")

	s.fetchBlocks(ctx)
	s.waitForSync.Done()

	if ctx.Err() != nil {
		return nil
	}

	sub, err := s.networkManager.PubSub().Subscribe(s.topic)
	if err != nil {
		return fmt.Errorf("Failed to subscribe to %s: %w", s.topic, err)
	}
	defer sub.Close()

	ch := sub.Start(ctx, true)
	for {
		select {
		case <-ctx.Done():
			s.logger.Debug().Msg("Syncer is terminated")
			return nil
		case data := <-ch:
			saved, err := s.processTopicTransaction(ctx, data)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to process topic transaction")
			}
			if !saved {
				s.fetchBlocks(ctx)
			}
		case <-time.After(s.config.Timeout):
			s.logger.Warn().Msgf("No new block in the topic for %s, pulling blocks actively", s.config.Timeout)

			s.fetchBlocks(ctx)
		}
	}
}

func (s *Syncer) processTopicTransaction(ctx context.Context, data []byte) (bool, error) {
	var pbBlock pb.RawFullBlock
	if err := proto.Unmarshal(data, &pbBlock); err != nil {
		return false, err
	}
	b, err := unmarshalBlockSSZ(&pbBlock)
	if err != nil {
		return false, err
	}

	block := b.Block
	s.logger.Debug().
		Stringer(logging.FieldBlockNumber, block.Id).
		Msg("Received block")

	if err := s.saveBlock(ctx, b); err != nil {
		switch {
		case errors.Is(err, errOutOfOrder):
			// todo: queue the block for later processing
			return false, nil
		case errors.Is(err, errOldBlock):
			return false, nil
		default:
			return false, err
		}
	}

	if uint64(block.Id)%uint64(blockReportInterval) == 0 {
		s.logger.Info().
			Uint64(logging.FieldBlockNumber, uint64(block.Id)).
			Stringer(logging.FieldBlockHash, block.Hash(s.config.ShardId)).
			Msg("Saved block")
	}

	return true, nil
}

func (s *Syncer) fetchBlocks(ctx context.Context) {
	// todo: fetch blocks until the queue (see todo above) is empty
	for {
		s.logger.Trace().Msg("Fetching next blocks")

		blocksCh := s.fetchBlocksRange(ctx)
		if blocksCh == nil {
			return
		}
		var count int
		for block := range blocksCh {
			count++
			if err := s.saveBlock(ctx, block); err != nil {
				if errors.Is(err, errOldBlock) {
					continue
				}
				s.logger.Error().
					Err(err).
					Stringer(logging.FieldBlockNumber, block.Id).
					Msg("Failed to save block")
				return
			}
		}
		if count == 0 {
			s.logger.Trace().Msg("No new blocks to fetch")
			return
		}
	}
}

func (s *Syncer) fetchBlocksRange(ctx context.Context) <-chan *types.BlockWithExtractedData {
	peers := ListPeers(s.networkManager, s.config.ShardId)

	if len(peers) == 0 {
		s.logger.Warn().Msg("No peers to fetch block from")
		return nil
	}

	s.logger.Trace().Msgf("Found %d peers to fetch block from:\n%v", len(peers), peers)

	lastBlock, _, err := s.validator.GetLastBlock(ctx)
	if err != nil {
		return nil
	}

	for _, p := range peers {
		s.logger.Trace().Msgf("Requesting blocks from %d from peer %s", lastBlock.Id+1, p)

		blocksCh, err := RequestBlocks(ctx, s.networkManager, p, s.config.ShardId, lastBlock.Id+1, s.logger)
		if err == nil {
			return blocksCh
		}

		if errors.As(err, &multistream.ErrNotSupported[network.ProtocolID]{}) {
			s.logger.Debug().Err(err).Msgf("Peer %s does not support the block protocol with our shard", p)
		} else {
			s.logger.Warn().Err(err).Msgf("Failed to request block from peer %s", p)
		}
	}

	return nil
}

func (s *Syncer) saveBlock(ctx context.Context, block *types.BlockWithExtractedData) error {
	if err := s.validator.ReplayBlock(ctx, block); err != nil {
		return err
	}
	s.notify(block.Id)

	s.logger.Trace().
		Stringer(logging.FieldBlockNumber, block.Block.Id).
		Msg("Block written")

	return nil
}

func (s *Syncer) GenerateZerostate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	if empty, err := s.shardIsEmpty(ctx); err != nil {
		return err
	} else if !empty {
		return nil
	}

	s.logger.Info().Msg("Generating zero-state...")
	gen, err := execution.NewBlockGenerator(ctx, s.config.BlockGeneratorParams, s.db, nil)
	if err != nil {
		return err
	}
	defer gen.Rollback()

	if _, err := gen.GenerateZeroState(s.config.ZeroStateConfig); err != nil {
		return err
	}
	return nil
}

func returnErrorOrPanic(err error) error {
	if assert.Enable {
		check.PanicIfErr(err)
	}
	return err
}
