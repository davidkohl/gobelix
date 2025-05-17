// internal/asxreader/reader.go
package asxreader

import (
	"fmt"
	"io"
	"time"

	"github.com/davidkohl/gobelix/asterix"
)

// AsterixReader provides a unified interface for reading ASTERIX messages
// regardless of the underlying transport protocol
type AsterixReader interface {
	io.Closer
	Next() (*asterix.DataBlock, error)
	Protocol() string
	Stats() ReaderStats
}

// ReaderStats contains statistics about the reader
type ReaderStats struct {
	BytesRead       int64
	MessagesRead    int64
	ConnectionTime  time.Duration
	SourceAddr      string // Remote address (if applicable)
	TransportErrors int    // Number of transport errors
	StartTime       time.Time
}

// NewReaderStats creates a new ReaderStats struct
func NewReaderStats() ReaderStats {
	return ReaderStats{
		StartTime: time.Now(),
	}
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
