package kyber

import (
	"go.dedis.ch/kyber/v3/pairing/bn256"
)

var suite *bn256.Suite

func init() {
	suite = bn256.NewSuite()
}
