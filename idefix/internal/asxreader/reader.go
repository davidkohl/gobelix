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

// DeadlineSetter is an interface for readers that support setting read deadlines
type DeadlineSetter interface {
	SetReadDeadline(t time.Time) error
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
	return NewAsterixReaderWithSkip(protocol, port, decoder, 0)
}

// NewAsterixReaderWithSkip creates an AsterixReader with optional skip bytes for UDP
func NewAsterixReaderWithSkip(protocol string, port int, decoder *asterix.Decoder, skipBytes int) (AsterixReader, error) {
	switch protocol {
	case "udp":
		return NewUDPAsterixReaderWithSkip(port, decoder, skipBytes)
	case "tcp":
		// TCP doesn't support skip bytes (use framing instead)
		if skipBytes > 0 {
			return nil, fmt.Errorf("skip-bytes not supported for TCP protocol")
		}
		return NewTCPAsterixReader(port, decoder)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
