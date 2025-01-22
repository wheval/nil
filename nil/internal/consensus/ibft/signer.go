package ibft

import (
	"crypto/ecdsa"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
	privateKey   *ecdsa.PrivateKey
	rawPublicKey []byte
}

func getHash(data []byte) []byte {
	return common.PoseidonHash(data).Bytes()
}

func NewSigner(privateKey *ecdsa.PrivateKey) *Signer {
	return &Signer{
		privateKey:   privateKey,
		rawPublicKey: crypto.CompressPubkey(&privateKey.PublicKey),
	}
}

func (s *Signer) SignHash(hash []byte) (types.Signature, error) {
	return crypto.Sign(hash, s.privateKey)
}

func (s *Signer) Sign(data []byte) (types.Signature, error) {
	return s.SignHash(getHash(data))
}

func (s *Signer) Verify(data []byte, signature types.Signature) bool {
	return s.VerifyWithKey(s.rawPublicKey, data, signature)
}

func (s *Signer) VerifyWithKey(publicKey []byte, data []byte, signature types.Signature) bool {
	return len(signature) >= 64 && crypto.VerifySignature(publicKey, getHash(data), signature[:64])
}

func (s *Signer) GetPublicKey() []byte {
	return s.rawPublicKey
}
