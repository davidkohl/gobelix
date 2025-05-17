// internal/asxreader/udp.go
package asxreader

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/davidkohl/gobelix/asterix"
)

// udpAsterixReader implements AsterixReader for UDP connections
type udpAsterixReader struct {
	conn      *net.UDPConn
	decoder   *asterix.Decoder
	stats     ReaderStats
	lastError error

	// For atomic access to stats
	bytesRead       int64
	messagesRead    int64
	transportErrors int32
}

// NewUDPAsterixReader creates a reader for UDP ASTERIX messages
func NewUDPAsterixReader(port int, decoder *asterix.Decoder) (AsterixReader, error) {
	if decoder == nil {
		return nil, fmt.Errorf("decoder cannot be nil")
	}

	// Create a specific UDP address to listen on
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Use ListenUDP directly
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
	}

	// Set initial read deadline to prevent blocking indefinitely
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	return &udpAsterixReader{
		conn:    conn,
		decoder: decoder,
		stats:   NewReaderStats(),
	}, nil
}

// Next reads and decodes the next ASTERIX message from UDP
func (r *udpAsterixReader) Next() (*asterix.DataBlock, error) {
	// Safety check
	if r.conn == nil {
		return nil, fmt.Errorf("nil UDP connection")
	}

	// Simple fixed buffer for UDP - no pool required
	buf := make([]byte, 65536) // Max UDP packet size

	// Read the next packet
	n, addr, err := r.conn.ReadFromUDP(buf)
	if err != nil {
		r.lastError = err
		atomic.AddInt32(&r.transportErrors, 1)

		// Check if it's a timeout error - this is expected when we have a read deadline
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("UDP read timeout: %w", err)
		}

		return nil, fmt.Errorf("reading UDP packet: %w", err)
	}

	// Handle empty packet
	if n == 0 {
		return nil, fmt.Errorf("received empty UDP packet")
	}

	// Update stats
	atomic.AddInt64(&r.bytesRead, int64(n))
	atomic.AddInt64(&r.messagesRead, 1)
	if addr != nil {
		r.stats.SourceAddr = addr.String()
	}
	r.stats.ConnectionTime = time.Since(r.stats.StartTime)

	// Use Decode instead of DecodeFrom since we already have the complete data
	msg, err := r.decoder.Decode(buf[:n])
	if err != nil {
		return nil, fmt.Errorf("decoding ASTERIX message: %w", err)
	}

	return msg, nil
}

// Close closes the underlying connection
func (r *udpAsterixReader) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Protocol returns the transport protocol name
func (r *udpAsterixReader) Protocol() string {
	return "UDP"
}

// Stats returns reader statistics
func (r *udpAsterixReader) Stats() ReaderStats {
	// Return a copy with atomic loads to avoid race conditions
	return ReaderStats{
		BytesRead:       atomic.LoadInt64(&r.bytesRead),
		MessagesRead:    atomic.LoadInt64(&r.messagesRead),
		TransportErrors: int(atomic.LoadInt32(&r.transportErrors)),
		ConnectionTime:  time.Since(r.stats.StartTime),
		SourceAddr:      r.stats.SourceAddr,
		StartTime:       r.stats.StartTime,
	}
}

// SetReadDeadline sets a deadline for the next ReadFromUDP call
func (r *udpAsterixReader) SetReadDeadline(t time.Time) error {
	if r.conn == nil {
		return fmt.Errorf("nil UDP connection")
	}
	return r.conn.SetReadDeadline(t)
}
