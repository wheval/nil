package ibft

import (
	"bytes"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/go-ibft/messages"
	protoIBFT "github.com/NilFoundation/nil/nil/go-ibft/messages/proto"
)

func (i *backendIBFT) IsValidProposal(rawProposal []byte) bool {
	proposal, err := i.unmarshalProposal(rawProposal)
	if err != nil {
		return false
	}

	_, err = i.scheduler.VerifyProposal(i.ctx, proposal)
	return err == nil
}

func (i *backendIBFT) IsValidValidator(msg *protoIBFT.IbftMessage) bool {
	msgNoSig, err := msg.PayloadNoSig()
	if err != nil {
		return false
	}

	if !i.signer.Verify(msgNoSig, msg.Signature) {
		event := i.logger.Error().Stringer(logging.FieldType, msg.GetType())
		if view := msg.GetView(); view != nil {
			event = event.Uint64(logging.FieldHeight, view.Height).
				Uint64(logging.FieldRound, view.Round)
		}
		event.Msg("Invalid signature")
		return false
	}

	return true
}

func (i *backendIBFT) IsProposer(id []byte, height, round uint64) bool {
	return true
}

func (i *backendIBFT) IsValidProposalHash(proposal *protoIBFT.Proposal, hash []byte) bool {
	prop, err := i.unmarshalProposal(proposal.RawProposal)
	if err != nil {
		return false
	}

	block, err := i.scheduler.VerifyProposal(i.ctx, prop)
	if err != nil {
		return false
	}

	blockHash := block.Hash(i.shardId)
	isValid := bytes.Equal(blockHash.Bytes(), hash)
	if !isValid {
		i.logger.Error().
			Stringer("expected", blockHash).
			Hex("got", hash).
			Uint64(logging.FieldRound, proposal.Round).
			Uint64(logging.FieldHeight, uint64(prop.PrevBlockId)+1).
			Msg("Invalid proposal hash")
	}
	return isValid
}

func (i *backendIBFT) IsValidCommittedSeal(
	proposalHash []byte,
	committedSeal *messages.CommittedSeal,
) bool {
	return true
}
