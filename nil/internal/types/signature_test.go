package types

import (
	"encoding/hex"
	"testing"

	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignature_MarshalText(t *testing.T) {
	t.Parallel()

	sig := Signature([]byte{0x01, 0x02, 0x03, 0x04})
	result, err := sig.MarshalText()
	require.NoError(t, err)

	expected := sig.String() // Converts to hex string with prefix "0x"
	assert.Equal(t, []byte(expected), result)
}

func TestSignature_UnmarshalText(t *testing.T) {
	t.Parallel()

	input := []byte("0x01020304")
	var sig Signature

	err := sig.UnmarshalText(input)
	require.NoError(t, err)

	expected := []byte{0x01, 0x02, 0x03, 0x04}
	assert.Equal(t, expected, []byte(sig))
}

func TestSignature_UnmarshalText_InvalidHex(t *testing.T) {
	t.Parallel()

	var sig Signature

	err := sig.UnmarshalText([]byte("invalid hex"))
	require.ErrorIs(t, err, hexutil.ErrMissingPrefix)

	err = sig.UnmarshalText([]byte("0xinvalidhex"))
	require.ErrorIs(t, err, hex.InvalidByteError('i'))
}
