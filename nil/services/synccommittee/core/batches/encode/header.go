package encode

import (
	"encoding/binary"
	"io"
)

const BatchMagic uint16 = 0xDEFA

type BatchHeader struct {
	Magic   uint16
	Version uint16
}

func NewBatchHeader(version uint16) BatchHeader {
	return BatchHeader{
		Magic:   BatchMagic,
		Version: version,
	}
}

func (bh BatchHeader) EncodeTo(out io.Writer) error {
	if err := binary.Write(out, binary.LittleEndian, bh.Magic); err != nil {
		return err
	}
	if err := binary.Write(out, binary.LittleEndian, bh.Version); err != nil {
		return err
	}
	return nil
}
