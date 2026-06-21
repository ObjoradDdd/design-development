package compressor

type Compressor interface {
	Compress(data []byte) ([]byte, error)
}
