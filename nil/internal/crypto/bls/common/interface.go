package common

// SecretKey represents a BLS secret or private key.
type PrivateKey interface {
	PublicKey() PublicKey
	Sign(msg []byte) (Signature, error)
	Marshal() ([]byte, error)
}

// PublicKey represents a BLS public key.
type PublicKey interface {
	Marshal() ([]byte, error)
}

// Mask reprensents the participants of aggregate public keys and signatures.
type Mask interface {
	SetParticipants(indices []uint32) error
	AggregatePublicKeys() (PublicKey, error)
}

// Signature represents a BLS signature.
type Signature interface {
	Verify(pubKey PublicKey, msg []byte) error
	Marshal() ([]byte, error)
}
