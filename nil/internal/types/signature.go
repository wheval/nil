package types

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/common/hexutil"
)

type Signature = hexutil.Bytes

type BlsSignature = hexutil.Bytes

type BlsAggregateSignature struct {
	Sig  hexutil.Bytes `json:"sig" yaml:"sig" ssz-max:"64"`
	Mask hexutil.Bytes `json:"mask" yaml:"mask" ssz-max:"128"`
}

func (b BlsAggregateSignature) String() string {
	return fmt.Sprintf("BlsAggregateSignature{Sig: %x, Mask: %x}", b.Sig, b.Mask)
}

// Generating this manually because fastssz purely supports nested structures
//go:generate go run github.com/NilFoundation/fastssz/sszgen --path signature.go -include ../../common/hexutil/bytes.go --objs BlsAggregateSignature
