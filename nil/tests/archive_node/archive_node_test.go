package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/stretchr/testify/suite"
)

type dbWrapper struct {
	db.DB

	Dropped bool
}

func (d *dbWrapper) Close() {
}

func (d *dbWrapper) DropAll() error {
	d.Dropped = true
	return d.DB.DropAll()
}

type SuiteArchiveNode struct {
	tests.ShardedSuite

	reusedDb *dbWrapper
	cancel   func()
	wg       sync.WaitGroup

	nShards            uint32
	withBootstrapPeers bool
	port               int
}

func (s *SuiteArchiveNode) SetupSuite() {
	s.nShards = 3

	reusedDb, err := db.NewBadgerDbInMemory()
	s.Require().NoError(err)
	s.reusedDb = &dbWrapper{DB: reusedDb}

	s.startCluster()

	s.startArchiveNode()
}

func (s *SuiteArchiveNode) startArchiveNode() {
	s.DbInit = func() db.DB {
		return s.reusedDb
	}

	ctx, cancel := context.WithCancel(s.Context)
	s.cancel = cancel
	s.DefaultClient, _ = s.StartArchiveNode(&tests.ArchiveNodeConfig{
		Ctx:                ctx,
		Wg:                 &s.wg,
		AllowDbDrop:        true,
		Port:               s.port + int(s.nShards),
		WithBootstrapPeers: s.withBootstrapPeers,
	})
}

func (s *SuiteArchiveNode) stopArchiveNode() {
	s.cancel()
	s.wg.Wait()
}

func (s *SuiteArchiveNode) startCluster() {
	s.DbInit = nil
	s.Start(&nilservice.Config{
		NShards:              s.nShards,
		CollatorTickPeriodMs: 200,
	}, s.port)
}

func (s *SuiteArchiveNode) TestRestarts() {
	check := func() {
		for shardId := range s.GetNShards() {
			b, err := s.DefaultClient.GetDebugBlock(s.Context, types.ShardId(shardId), 0, true)
			s.Require().NoError(err)
			s.NotNil(b)
		}

		for shardId := range s.GetNShards() {
			s.Require().Eventually(func() bool {
				b, err := s.DefaultClient.GetDebugBlock(s.Context, types.ShardId(shardId), 1, true)
				s.Require().NoError(err)
				return b != nil
			}, 5*time.Second, 100*time.Millisecond)
		}
	}

	s.Run("Before", check)

	s.Run("Restart", func() {
		s.stopArchiveNode()
		s.startArchiveNode()

		s.False(s.reusedDb.Dropped)
	})

	s.Run("AfterRestart", check)

	if !s.withBootstrapPeers {
		// Bootstrap peers are required for the node to fetch the version.
		// todo: fix it
		return
	}

	s.Run("ClusterReset", func() {
		s.Cancel()
		s.startCluster()
		s.startArchiveNode()

		s.True(s.reusedDb.Dropped)
	})

	s.Run("AfterClusterReset", check)
}

func (s *SuiteArchiveNode) TestGetFaucetBalance() {
	value, err := s.DefaultClient.GetBalance(s.Context, types.FaucetAddress, "latest")
	s.Require().NoError(err)
	s.Positive(value.Uint64())
}

func TestArchiveNodeWithBootstrapPeers(t *testing.T) {
	t.Parallel()

	suite.Run(t, &SuiteArchiveNode{
		withBootstrapPeers: true,
		port:               10005,
	})
}

func TestArchiveNodeWithoutBootstrapPeers(t *testing.T) {
	t.Parallel()

	suite.Run(t, &SuiteArchiveNode{
		withBootstrapPeers: false,
		port:               10015,
	})
}
