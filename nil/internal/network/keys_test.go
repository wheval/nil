package network

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testLoadOrGenerateKeys[T any](t *testing.T, loadOrGenerateFunc func(string) (T, error)) {
	t.Helper()

	tempDir := t.TempDir()
	fileName := tempDir + "/keys.yaml"

	privKey, err := loadOrGenerateFunc(fileName)
	require.NoError(t, err)

	t.Run("load", func(t *testing.T) {
		t.Parallel()

		loaded, err := loadOrGenerateFunc(fileName)
		require.NoError(t, err)

		require.Equal(t, privKey, loaded)
	})

	t.Run("new file", func(t *testing.T) {
		t.Parallel()

		newFileName := tempDir + "/new-keys.yaml"
		require.NotEqual(t, fileName, newFileName)

		generated, err := loadOrGenerateFunc(newFileName)
		require.NoError(t, err)

		require.NotEqual(t, privKey, generated)
	})
}

func TestLoadOrGenerateKeys(t *testing.T) {
	t.Parallel()
	testLoadOrGenerateKeys(t, LoadOrGenerateKeys)
}
