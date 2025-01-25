package cliservice

import (
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/types"
)

func (s *Service) handleTokenTx(txHash common.Hash, contractAddr types.Address) error {
	s.logger.Info().
		Stringer(logging.FieldShardId, contractAddr.ShardId()).
		Stringer(logging.FieldTransactionHash, txHash).
		Send()

	_, err := s.WaitForReceipt(txHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to wait for token transaction receipt")
		return err
	}
	return nil
}

func (s *Service) TokenCreate(contractAddr types.Address, amount types.Value, name string) (*types.TokenId, error) {
	txHash, err := s.client.SetTokenName(s.ctx, contractAddr, name, s.privateKey)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to send setTokenName transaction")
		return nil, err
	}
	if err = s.handleTokenTx(txHash, contractAddr); err != nil {
		return nil, err
	}

	txHash, err = s.client.ChangeTokenAmount(s.ctx, contractAddr, amount, s.privateKey, true /* mint */)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to send minToken transaction")
		return nil, err
	}
	if err = s.handleTokenTx(txHash, contractAddr); err != nil {
		return nil, err
	}

	tokenId := types.TokenIdForAddress(contractAddr)
	s.logger.Info().Stringer(logging.FieldTokenId, common.BytesToHash(tokenId[:])).Msgf("Created %v:%v", name, amount)
	return tokenId, nil
}

func (s *Service) ChangeTokenAmount(contractAddr types.Address, amount types.Value, mint bool) (common.Hash, error) {
	txHash, err := s.client.ChangeTokenAmount(s.ctx, contractAddr, amount, s.privateKey, mint)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to send transaction for token amount change")
		return common.EmptyHash, err
	}

	if err = s.handleTokenTx(txHash, contractAddr); err != nil {
		return common.EmptyHash, err
	}
	tokenId := types.TokenIdForAddress(contractAddr)
	operation := "Minted"
	if !mint {
		operation = "Burned"
	}
	s.logger.Info().Stringer(logging.FieldTokenId, common.BytesToHash(tokenId[:])).Msgf("%s %v", operation, amount)
	return txHash, nil
}
