package network

import (
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func mustDecodePeerId(t *testing.T, s string) peer.ID {
	t.Helper()

	id, err := peer.Decode(s)
	require.NoError(t, err)
	return id
}

func mustDecodeMultiaddr(t *testing.T, s string) ma.Multiaddr {
	t.Helper()
	addr, err := ma.NewMultiaddr(s)
	require.NoError(t, err)
	return addr
}

func TestConfigYamlSerialization(t *testing.T) {
	t.Parallel()

	addrInfo1 := AddrInfo{
		ID:    mustDecodePeerId(t, "16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf"),
		Addrs: []ma.Multiaddr{mustDecodeMultiaddr(t, "/ip4/127.0.0.1/tcp/1500")},
	}

	addrInfo2 := AddrInfo{
		ID: mustDecodePeerId(t, "16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg"),
		Addrs: []ma.Multiaddr{
			mustDecodeMultiaddr(t, "/ip4/192.168.1.1/tcp/4002"),
			mustDecodeMultiaddr(t, "/ip4/127.0.0.1/tcp/4002"),
		},
	}

	config := Config{
		DHTBootstrapPeers:  AsAddrInfoSlice(addrInfo1, addrInfo2),
		RelayPublicAddress: addrInfo1,
	}

	t.Run("marshal unmarshal", func(t *testing.T) {
		t.Parallel()

		data, err := yaml.Marshal(config)
		require.NoError(t, err)

		expectedYaml := `---
dhtBootstrapPeers:
  - /ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf
  - /ip4/192.168.1.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg
  - /ip4/127.0.0.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg
relayPublicAddress: /ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf
`
		require.YAMLEq(t, expectedYaml, string(data))

		var deserialized Config
		err = yaml.Unmarshal(data, &deserialized)
		require.NoError(t, err)

		require.Equal(t, config, deserialized)
	})

	t.Run("unmarshal unordered by peer id", func(t *testing.T) {
		t.Parallel()

		configYaml := `---
dhtBootstrapPeers:
  - /ip4/192.168.1.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg
  - /ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf
  - /ip4/127.0.0.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg
relayPublicAddress: /ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf
`

		var deserialized Config
		err := yaml.Unmarshal([]byte(configYaml), &deserialized)
		require.NoError(t, err)

		require.Equal(t, config, deserialized)
	})

	t.Run("unmarshal invalid not sequence but scalar", func(t *testing.T) {
		t.Parallel()

		var deserialized Config
		invalidYaml := `---
dhtBootstrapPeers: /ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf
`
		err := yaml.Unmarshal([]byte(invalidYaml), &deserialized)
		require.Error(t, err)
	})

	t.Run("unmarshal invelid not sequence but comma separated value", func(t *testing.T) {
		t.Parallel()

		var deserialized Config
		invalidYaml := "dhtBootstrapPeers: " +
			"/ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf," +
			"/ip4/192.168.1.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg"
		err := yaml.Unmarshal([]byte(invalidYaml), &deserialized)
		require.Error(t, err)
	})

	t.Run("unmarshal multiaddress", func(t *testing.T) {
		t.Parallel()

		var deserialized Config
		configYaml := "relayPublicAddress: " +
			"/ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf," +
			"/ip4/192.168.1.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf"

		err := yaml.Unmarshal([]byte(configYaml), &deserialized)
		require.NoError(t, err)
	})

	t.Run("unmarshal invalid multiaddress", func(t *testing.T) {
		t.Parallel()

		var deserialized Config
		configYaml := "relayPublicAddress: " +
			"/ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf," +
			"/ip4/192.168.1.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg"

		err := yaml.Unmarshal([]byte(configYaml), &deserialized)
		require.ErrorContains(t, err, "not all multiaddresses belong to the same peer")
	})
}

func TestAddrInfoStringRepresentation(t *testing.T) {
	t.Parallel()

	addr1 := "/ip4/127.0.0.1/tcp/1500/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf"
	addr2 := "/ip4/192.168.0.10/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yf"

	transport1, id1 := peer.SplitAddr(mustDecodeMultiaddr(t, addr1))
	transport2, id2 := peer.SplitAddr(mustDecodeMultiaddr(t, addr2))
	require.Equal(t, id1, id2)
	addrInfo1 := AddrInfo{
		ID:    id1,
		Addrs: []ma.Multiaddr{transport1, transport2},
	}

	t.Run("marshal unmarshal", func(t *testing.T) {
		t.Parallel()

		text, err := addrInfo1.MarshalText()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s,%s", addr1, addr2), string(text))

		parsedAddrInfo := AddrInfo{}
		err = parsedAddrInfo.UnmarshalText(text)
		require.NoError(t, err)
		require.Equal(t, addrInfo1, parsedAddrInfo)
	})

	addr3 := "/ip4/127.0.0.1/tcp/4002/p2p/16Uiu2HAmQctkUi9y7WfUtYa9rPon1m5TRBtXSvUwi2VtpbWZj4yg"
	transport3, id3 := peer.SplitAddr(mustDecodeMultiaddr(t, addr3))
	addrInfo2 := AddrInfo{
		ID:    id3,
		Addrs: []ma.Multiaddr{transport3},
	}
	addrInfos := AsAddrInfoSlice(addrInfo1, addrInfo2)

	t.Run("marshal unmarshal two infos", func(t *testing.T) {
		t.Parallel()

		text, err := addrInfos.MarshalText()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s,%s,%s", addr1, addr2, addr3), string(text))
	})

	t.Run("marshal unmarshal unordered by peer id", func(t *testing.T) {
		t.Parallel()

		text := []byte(fmt.Sprintf("%s,%s,%s", addr1, addr3, addr2))
		var parsedAddrInfos AddrInfoSlice
		err := parsedAddrInfos.UnmarshalText(text)
		require.NoError(t, err)
		require.Equal(t, addrInfos, parsedAddrInfos)
	})
}
