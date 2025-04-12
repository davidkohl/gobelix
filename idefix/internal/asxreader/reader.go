package asxreader

// reader.go

import (
	"fmt"
	"io"

	"github.com/davidkohl/gobelix/asterix"
)

// AsterixReader provides a unified interface for reading ASTERIX messages
// regardless of the underlying transport protocol
type AsterixReader interface {
	io.Closer
	Next() (*asterix.AsterixMessage, error)
	Protocol() string
}

// NewAsterixReader creates an appropriate AsterixReader based on protocol
func NewAsterixReader(protocol string, port int, decoder *asterix.Decoder) (AsterixReader, error) {
	switch protocol {
	case "udp":
		return NewUDPAsterixReader(port, decoder)
	case "tcp":
		return NewTCPAsterixReader(port, decoder)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
