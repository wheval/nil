package keys

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/NilFoundation/nil/nil/common/logging"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v3"
)

var Logger = logging.NewLogger("keys")

var errKeysNotInitialized = errors.New("keys are not initialized")

type dumpedValidatorKey struct {
	PrivateKey hexutil.Bytes `yaml:"privateKey"`
	PublicKey  hexutil.Bytes `yaml:"publicKey"`
}

type ValidatorKeysManager struct {
	validatorKeyPath string
	key              *ecdsa.PrivateKey
}

func NewValidatorKeyManager(validatorKeyPath string) *ValidatorKeysManager {
	return &ValidatorKeysManager{
		validatorKeyPath: validatorKeyPath,
	}
}

func (v *ValidatorKeysManager) generateKey() error {
	var err error
	v.key, err = ecdsa.GenerateKey(gethcrypto.S256(), rand.Reader)
	if err != nil {
		return err
	}
	return nil
}

const filePermissions = 0o644

func (v *ValidatorKeysManager) dumpKey() error {
	dumpedKey := &dumpedValidatorKey{
		PrivateKey: gethcrypto.FromECDSA(v.key),
		PublicKey:  gethcrypto.FromECDSAPub(&v.key.PublicKey),
	}

	data, err := yaml.Marshal(dumpedKey)
	if err != nil {
		return err
	}
	return os.WriteFile(v.validatorKeyPath, data, filePermissions)
}

func (v *ValidatorKeysManager) loadKey() error {
	Logger.Info().Msgf("Loading key from path: %s", v.validatorKeyPath)
	data, err := os.ReadFile(v.validatorKeyPath)
	if err != nil {
		return err
	}

	dumpedKey := &dumpedValidatorKey{}
	if err := yaml.Unmarshal(data, dumpedKey); err != nil {
		return err
	}

	privKey, err := gethcrypto.ToECDSA(dumpedKey.PrivateKey)
	if err != nil {
		return err
	}
	pubKey, err := gethcrypto.UnmarshalPubkey(dumpedKey.PublicKey)
	if err != nil {
		return err
	}
	if !pubKey.Equal(&privKey.PublicKey) {
		return errors.New("public key mismatch")
	}
	v.key = privKey
	return nil
}

// This functions initializes key by loading it from the file if it exists,
// or generating new key and saving it to the file.
func (v *ValidatorKeysManager) InitKey() error {
	if v.key != nil {
		return errors.New("key is already initialized")
	}
	if _, err := os.Stat(v.validatorKeyPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Error checking key file: %w", err)
		}
		Logger.Warn().Msgf("Key file not found, generating new key at path: %s", v.validatorKeyPath)
		if err := v.generateKey(); err != nil {
			return fmt.Errorf("Error generating key: %w", err)
		}
		if err := v.dumpKey(); err != nil {
			return fmt.Errorf("Error saving key: %w", err)
		}
		return nil
	}
	if err := v.loadKey(); err != nil {
		return fmt.Errorf("Error loading key: %w", err)
	}
	return nil
}

func (v *ValidatorKeysManager) GetKey() (*ecdsa.PrivateKey, error) {
	if v.key == nil {
		return nil, errKeysNotInitialized
	}
	return v.key, nil
}

func (v *ValidatorKeysManager) GetPublicKey() ([]byte, error) {
	if v.key == nil {
		return nil, errKeysNotInitialized
	}
	return gethcrypto.CompressPubkey(&v.key.PublicKey), nil
}

func (v *ValidatorKeysManager) GetKeysPath() string {
	return v.validatorKeyPath
}
