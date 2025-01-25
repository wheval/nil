package keys

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitKeys(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fileName := tempDir + "/keys.yaml"
	const nShards = 5
	validatorKeysManager := NewValidatorKeyManager(fileName, nShards)
	require.NoError(t, validatorKeysManager.InitKeys())

	keys, err := validatorKeysManager.GetKeys()
	require.NoError(t, err)

	validatorKeysManager2 := NewValidatorKeyManager(fileName, nShards)
	require.NoError(t, validatorKeysManager2.InitKeys())

	keys2, err := validatorKeysManager2.GetKeys()
	require.NoError(t, err)

	require.Equal(t, keys, keys2)
}
