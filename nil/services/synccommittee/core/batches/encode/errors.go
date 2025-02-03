package encode

import "errors"

var (
	ErrInvalidMagic   = errors.New("invalid_batch_magic")
	ErrInvalidVersion = errors.New("invalid_batch_encoding_version")
)
