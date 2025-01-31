package keys

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitKeys(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fileName := tempDir + "/keys.yaml"
	validatorKeysManager := NewValidatorKeyManager(fileName)
	require.NoError(t, validatorKeysManager.InitKey())

	keys, err := validatorKeysManager.GetKey()
	require.NoError(t, err)

	validatorKeysManager2 := NewValidatorKeyManager(fileName)
	require.NoError(t, validatorKeysManager2.InitKey())

	keys2, err := validatorKeysManager2.GetKey()
	require.NoError(t, err)

	require.Equal(t, keys, keys2)
}
