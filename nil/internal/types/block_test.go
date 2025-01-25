package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlock_SignAndVerifySignature(t *testing.T) {
	t.Parallel()

	block := &Block{}

	// Generate keys
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubKey := crypto.FromECDSAPub(&privKey.PublicKey)

	// Sign the block
	err = block.Sign(privKey, BaseShardId)
	require.NoError(t, err)
	assert.NotEmpty(t, block.Signature)

	// Check signature
	err = block.VerifySignature(pubKey, BaseShardId)
	require.NoError(t, err)

	// Invalid public key
	invalidPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	invalidPubKey := crypto.FromECDSAPub(&invalidPrivKey.PublicKey)

	err = block.VerifySignature(invalidPubKey, BaseShardId)
	require.EqualError(t, err, "invalid signature")

	// Empty public key
	err = block.VerifySignature(nil, BaseShardId)
	require.EqualError(t, err, "invalid signature")

	// Attempt to sign twice
	err = block.Sign(privKey, BaseShardId)
	require.EqualError(t, err, "block is already signed")

	// Verify with empty signature
	block.Signature = nil
	err = block.VerifySignature(nil, BaseShardId)
	require.EqualError(t, err, "invalid signature")
}
