package cliservice

import (
	"encoding/json"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
)

// FetchReceiptByHash fetches the transaction receipt by hash
func (s *Service) FetchReceiptByHash(hash common.Hash) (*jsonrpc.RPCReceipt, error) {
	receiptData, err := s.client.GetInTransactionReceipt(s.ctx, hash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to fetch receipt")
		return nil, err
	}
	return receiptData, nil
}

// FetchReceiptByHashJson fetches the transaction receipt as a JSON string
func (s *Service) FetchReceiptByHashJson(hash common.Hash) ([]byte, error) {
	receipt, err := s.FetchReceiptByHash(hash)
	if err != nil {
		return nil, err
	}
	receiptDataJSON, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal receipt data to JSON")
		return nil, err
	}
	return receiptDataJSON, nil
}
