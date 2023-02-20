package encoding

import "io"

// Codec format data to []byte, decode data from []byte
type Codec interface {
	// Marshal format data to []byte
	Marshal(interface{}) ([]byte, error)
	// Unmarshal decode data from []byte
	Unmarshal([]byte, interface{}) error
}

// Compressor compressor interface
type Compressor interface {
	// Compress get compress writer
	Compress(io.Writer) (io.WriteCloser, error)
	// Decompress get decompress reader
	Decompress(io.Reader) (io.ReadCloser, error)
	// SetLevel set compress level
	SetLevel(int) error
}
