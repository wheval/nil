package network

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type AddrInfo peer.AddrInfo

var _ pflag.Value = (*AddrInfo)(nil)

func (a AddrInfo) Empty() bool {
	return errors.Is(a.ID.Validate(), peer.ErrEmptyPeerID)
}

func (a *AddrInfo) Set(value string) error {
	addrInfos, err := addrInfosFromString(value)
	if err != nil {
		return err
	}
	if len(addrInfos) == 0 {
		return fmt.Errorf("%s is not a valid address", value)
	}
	if len(addrInfos) > 1 {
		return fmt.Errorf("not all multiaddresses belong to the same peer: %s", value)
	}
	*a = AddrInfo(addrInfos[0])
	return nil
}

func (a AddrInfo) String() string {
	return addrInfosToString(peer.AddrInfo(a))
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

type AddrInfoSlice []peer.AddrInfo

func AsAddrInfoSlice(addrInfos ...AddrInfo) AddrInfoSlice {
	return slices.Collect(
		common.Transform(
			slices.Values(addrInfos),
			func(a AddrInfo) peer.AddrInfo { return peer.AddrInfo(a) }))
}

var _ pflag.Value = (*AddrInfoSlice)(nil)

func (s *AddrInfoSlice) Set(value string) (err error) {
	*s, err = addrInfosFromString(value)
	return
}

// String returns one multiaddress per item including PeerID, comma-separated
func (s *AddrInfoSlice) String() string {
	return addrInfosToString(*s...)
}

func (s *AddrInfoSlice) Type() string {
	return "[]AddrInfo"
}

// Strings returns one multiaddress per item including PeerID, in each array element
func (s *AddrInfoSlice) Strings() ([]string, error) {
	return addrInfosToStrings(*s...)
}

func (s *AddrInfoSlice) FromStrings(addrs []string) (err error) {
	*s, err = addrInfosFromStrings(addrs)
	return
}

func (s AddrInfoSlice) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *AddrInfoSlice) UnmarshalText(text []byte) (err error) {
	return s.Set(string(text))
}

func (s AddrInfoSlice) MarshalYAML() (any, error) {
	return addrInfosToStrings(s...)
}

func (s *AddrInfoSlice) UnmarshalYAML(node *yaml.Node) error {
	var addrs []string
	if err := node.Decode(&addrs); err != nil {
		return err
	}
	return s.FromStrings(addrs)
}

func addrInfosFromStrings(addrs []string) ([]peer.AddrInfo, error) {
	var err error
	mas := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		mas[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
	}

	addrInfos, err := peer.AddrInfosFromP2pAddrs(mas...)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(addrInfos, func(l, r peer.AddrInfo) int {
		return strings.Compare(string(l.ID), string(r.ID))
	})
	return addrInfos, nil
}

func addrInfosFromString(value string) ([]peer.AddrInfo, error) {
	addrs, err := common.ReadAsCSV(value)
	if err != nil {
		return nil, err
	}
	return addrInfosFromStrings(addrs)
}

func addrInfosToStrings(addrInfos ...peer.AddrInfo) ([]string, error) {
	result := make([]string, 0, len(addrInfos))
	for _, addrInfo := range addrInfos {
		multiAddrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
		if err != nil {
			return nil, err
		}
		result = slices.AppendSeq(
			result,
			common.Transform(
				slices.Values(multiAddrs),
				func(addr ma.Multiaddr) string { return addr.String() }))
	}
	return result, nil
}

func addrInfosToString(addrInfos ...peer.AddrInfo) string {
	addrs, err := addrInfosToStrings(addrInfos...)
	if err != nil {
		return err.Error()
	}
	result, err := common.WriteAsCSV(addrs)
	if err != nil {
		return err.Error()
	}
	return result
}
