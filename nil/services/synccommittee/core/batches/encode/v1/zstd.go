package v1

import (
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog"
)

// some aux interfaces for operating on memory buffers
type LenAware interface {
	Len() int
}

type Growable interface {
	Grow(int)
}

type zstdCompressor struct {
	logger zerolog.Logger
}

func NewZstdCompressor(logger zerolog.Logger) *zstdCompressor {
	return &zstdCompressor{
		logger: logger,
	}
}

func (zc *zstdCompressor) Compress(from io.Reader, to io.Writer) (err error) {
	impl, err := zstd.NewWriter(to, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
	if err != nil {
		return err
	}

	defer func() {
		closeErr := impl.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var maxEstimatedSize int
	if lenAware, ok := from.(LenAware); ok {
		maxEstimatedSize = lenAware.Len()
		zc.logger.Trace().Int("estimated_size", maxEstimatedSize).Send()
	}

	// adjust the sizes if we are operating with memory buffers
	if growable, ok := to.(Growable); maxEstimatedSize > 0 && ok {
		growable.Grow(maxEstimatedSize)
	}

	_, err = impl.ReadFrom(from)
	return
}
