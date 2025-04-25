package network

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/suite"
)

type RelayTestSuite struct {
	networkSuite
}

func (s *RelayTestSuite) SetupSuite() {
	s.port = 1678
}

func (s *RelayTestSuite) TestRelay() {
	// Forcing public reachability, otherwise the relay service will not start
	relay := s.newManagerWithBaseConfig(&Config{
		ServeRelay:   true,
		Reachability: network.ReachabilityPublic,
	})
	defer relay.Close()

	// Forcing private reachability, otherwise the private node won't use the relay
	private := s.newManagerWithBaseConfig(&Config{
		Relays:       AsAddrInfoSlice(CalcAddress(relay)),
		Reachability: network.ReachabilityPrivate,
	})
	private.SetRequestHandler(s.context, "/hello", func(context.Context, []byte) ([]byte, error) {
		return []byte("world"), nil
	})
	defer private.Close()

	// Connect the private node to the relay (avoiding discovery)
	ConnectManagers(s.T(), private, relay)

	// The client node must be able to connect to the private node via the relay
	client := s.newManager()
	defer client.Close()

	relayedAddr := RelayedAddress(private.host.ID(), CalcAddress(relay))
	id, err := client.Connect(s.context, relayedAddr)
	s.Require().NoError(err)
	s.Require().Equal(private.host.ID(), id)

	resp, err := client.SendRequestAndGetResponse(
		network.WithAllowLimitedConn(s.context, "relay"), id, "/hello", []byte("hello"))
	s.Require().NoError(err)
	s.Require().Equal("world", string(resp))
}

func TestRelay(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(RelayTestSuite))
}
