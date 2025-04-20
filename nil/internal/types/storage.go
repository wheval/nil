package types

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/common"
)

type Storage map[common.Hash]Uint256

func (s Storage) String() (str string) {
	for key, value := range s {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (s Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range s {
		cpy[key] = value
	}

	return cpy
}
