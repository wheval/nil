package cliservice

import (
	"encoding/json"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
)

// FetchTransactionByHashJson fetches the transaction by hash
func (s *Service) FetchTransactionByHashJson(hash common.Hash) ([]byte, error) {
	transactionData, err := s.FetchTransactionByHash(hash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to fetch transaction")
		return nil, err
	}

	transactionDataJSON, err := json.MarshalIndent(transactionData, "", "  ")
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal transaction data to JSON")
		return nil, err
	}

	s.logger.Info().Msgf("Fetched transaction:\n%s", transactionDataJSON)
	return transactionDataJSON, nil
}

func (s *Service) FetchTransactionByHash(hash common.Hash) (*jsonrpc.RPCInTransaction, error) {
	return s.client.GetInTransactionByHash(s.ctx, hash)
}
