package kyber

import (
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/internal/crypto/bls/common"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

// Implement PublicKey interface
type PublicKey struct{ p kyberPublicKey }

var _ common.PublicKey = (*PublicKey)(nil)

func (pk *PublicKey) Marshal() ([]byte, error) {
	return pk.p.MarshalBinary()
}

// Implement Mask interface
type Mask struct{ m *kyberMask }

var _ common.Mask = (*Mask)(nil)

func (m *Mask) SetParticipants(indices []uint32) error {
	mask := make([]byte, m.m.Len())
	if err := m.m.SetMask(mask); err != nil {
		return err
	}
	for _, i := range indices {
		if err := m.m.SetBit(int(i), true); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mask) AggregatePublicKeys() (common.PublicKey, error) {
	pk, err := bdn.AggregatePublicKeys(suite, m.m)
	if err != nil {
		return nil, err
	}
	return &PublicKey{p: pk}, nil
}

func NewMask(publics []common.PublicKey) (*Mask, error) {
	kyberPublics := make([]kyberPublicKey, len(publics))
	for i, p := range publics {
		p, ok := p.(*PublicKey)
		check.PanicIfNot(ok)
		kyberPublics[i] = p.p
	}

	mask, err := sign.NewMask(suite, kyberPublics, nil)
	if err != nil {
		return nil, err
	}
	return &Mask{mask}, nil
}

// Implement Signature interface
type Signature struct {
	s []byte
}

var _ common.Signature = (*Signature)(nil)

func (s *Signature) Marshal() ([]byte, error) {
	return s.s, nil
}

func (s *Signature) Verify(pubKey common.PublicKey, msg []byte) error {
	p, ok := pubKey.(*PublicKey)
	check.PanicIfNot(ok)
	return bdn.Verify(suite, p.p, msg, s.s)
}

// Implement PrivateKey interface
type PrivateKey struct {
	s kyberPrivateKey
	p PublicKey
}

var _ common.PrivateKey = (*PrivateKey)(nil)

func (sk *PrivateKey) PublicKey() common.PublicKey {
	return &sk.p
}

func (sk *PrivateKey) Sign(msg []byte) (common.Signature, error) {
	sig, err := bdn.Sign(suite, sk.s, msg)
	if err != nil {
		return nil, err
	}
	return &Signature{s: sig}, nil
}

func (sk *PrivateKey) Marshal() ([]byte, error) {
	return sk.s.MarshalBinary()
}

func NewRandomKey() common.PrivateKey {
	sk, pk := bdn.NewKeyPair(suite, suite.RandomStream())
	return &PrivateKey{s: sk, p: PublicKey{pk}}
}

func AggregateSignatures(sigs []common.Signature, mask common.Mask) (common.Signature, error) {
	kyberSigs := make([][]byte, len(sigs))
	for i, s := range sigs {
		s, ok := s.(*Signature)
		check.PanicIfNot(ok)
		kyberSigs[i] = s.s
	}

	m, ok := mask.(*Mask)
	check.PanicIfNot(ok)
	sig, err := bdn.AggregateSignatures(suite, kyberSigs, m.m)
	if err != nil {
		return nil, err
	}

	s, err := sig.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return &Signature{s: s}, nil
}
