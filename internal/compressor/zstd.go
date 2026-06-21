package compressor

import (
	"github.com/klauspost/compress/zstd"
)

type ZstdCompressor struct {
	encoder *zstd.Encoder
}

func NewZstdCompressor(level int) (*ZstdCompressor, error) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
	if err != nil {
		return nil, err
	}

	return &ZstdCompressor{
		encoder: enc,
	}, nil
}

func (z *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	compressed := z.encoder.EncodeAll(data, nil)

	return compressed, nil
}
