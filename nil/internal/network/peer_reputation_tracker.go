package network

import cm "github.com/NilFoundation/nil/nil/internal/network/connection_manager"

func TryGetPeerReputationTracker(manager *BasicManager) cm.PeerReputationTracker {
	return cm.TryGetPeerReputationTracker(manager.host)
}
