package compressor

import (
	"bytes"
	"sync"

	"github.com/klauspost/compress/gzip"
)

type GzipCompressor struct {
	pool sync.Pool
}

func NewGzipCompressor(level int) *GzipCompressor {
	return &GzipCompressor{
		pool: sync.Pool{
			New: func() any {
				writer, _ := gzip.NewWriterLevel(nil, level)
				return writer
			},
		},
	}
}

func (g *GzipCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	writer := g.pool.Get().(*gzip.Writer)
	defer g.pool.Put(writer)

	writer.Reset(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}