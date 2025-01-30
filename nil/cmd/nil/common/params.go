package common

import "github.com/NilFoundation/nil/nil/internal/types"

type Params struct {
	AbiPath          string
	WithDetails      bool
	AsJson           bool
	Fee              types.FeePack
	InOverridesPath  string
	OutOverridesPath string
}

var Quiet = false
