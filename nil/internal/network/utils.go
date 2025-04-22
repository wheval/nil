package network

import (
	"github.com/NilFoundation/nil/nil/common/check"
	ma "github.com/multiformats/go-multiaddr"
)

func CalcAddress(m Manager) AddrInfo {
	return AddrInfo{m.getHost().ID(), m.getHost().Addrs()}
}

func RelayedAddress(p PeerID, relayAddrInfo AddrInfo) AddrInfo {
	addrInfo := AddrInfo{
		ID:    p,
		Addrs: make([]ma.Multiaddr, len(relayAddrInfo.Addrs)),
	}
	for i, relayAddr := range relayAddrInfo.Addrs {
		var addr ma.Multiaddr
		relay, err := ma.NewComponent("p2p", relayAddrInfo.ID.String())
		check.PanicIfErr(err)
		addr = addr.AppendComponent(relay)
		circuit, err := ma.NewComponent("p2p-circuit", "")
		check.PanicIfErr(err)
		addr = addr.AppendComponent(circuit)
		addrInfo.Addrs[i] = relayAddr.Encapsulate(addr)
	}
	return addrInfo
}
