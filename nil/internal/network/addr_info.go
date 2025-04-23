package network

import (
	"errors"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"gopkg.in/yaml.v3"
)

type AddrInfo peer.AddrInfo

func (a AddrInfo) Empty() bool {
	return errors.Is(a.ID.Validate(), peer.ErrEmptyPeerID)
}

func (a *AddrInfo) Set(value string) error {
	addr, err := peer.AddrInfoFromString(value)
	if err != nil {
		return err
	}
	*a = AddrInfo(*addr)
	return nil
}

func (a AddrInfo) String() string {
	mu, err := peer.AddrInfoToP2pAddrs((*peer.AddrInfo)(&a))
	if err != nil {
		return err.Error()
	}
	values := make([]string, len(mu))
	for i, val := range mu {
		values[i] = val.String()
	}
	str, err := common.WriteAsCSV(values)
	if err != nil {
		return err.Error()
	}
	return str
}

func (a *AddrInfo) Type() string {
	return "AddrInfo"
}

func (a AddrInfo) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *AddrInfo) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

func (a *AddrInfo) MarshalYAML() (any, error) {
	return a.MarshalText()
}

func (a *AddrInfo) UnmarshalYAML(value *yaml.Node) error {
	return a.UnmarshalText([]byte(value.Value))
}

type AddrInfoSlice = common.PValueSlice[*AddrInfo, AddrInfo]

func ToLibP2pAddrInfoSlice(s AddrInfoSlice) []peer.AddrInfo {
	res := make([]peer.AddrInfo, len(s))
	for i, a := range s {
		res[i] = peer.AddrInfo(a)
	}
	return res
}
