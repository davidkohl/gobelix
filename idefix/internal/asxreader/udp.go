// internal/asxreader/udp.go
package net

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/encoding"
)

// udpAsterixReader implements AsterixReader for UDP connections
type udpAsterixReader struct {
	conn       net.PacketConn
	decoder    *encoding.Decoder
	stats      ReaderStats
	bufferPool *encoding.BufferPool
	lastError  error

	// For atomic access to stats
	bytesRead       int64
	messagesRead    int64
	transportErrors int64
}

// NewUDPAsterixReader creates a reader for UDP ASTERIX messages
func NewUDPAsterixReader(port int, decoder *encoding.Decoder) (AsterixReader, error) {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
	}

	// Create a buffer pool for better memory management
	pool := encoding.NewBufferPool()

	return &udpAsterixReader{
		conn:       conn,
		decoder:    decoder,
		stats:      NewReaderStats(),
		bufferPool: pool,
	}, nil
}

// Next reads and decodes the next ASTERIX message from UDP
func (r *udpAsterixReader) Next() (*asterix.AsterixMessage, error) {
	// Get a buffer from the pool
	buf := r.bufferPool.Get(65536) // Max UDP packet size
	defer r.bufferPool.Put(buf)

	// Set read deadline if needed (e.g., to support timeout)
	// r.conn.SetReadDeadline(time.Now().Add(readTimeout))

	// Read the next packet
	n, addr, err := r.conn.ReadFrom(buf)
	if err != nil {
		r.lastError = err
		atomic.AddInt64(&r.transportErrors, 1)
		return nil, fmt.Errorf("reading UDP packet: %w", err)
	}

	// Update stats
	atomic.AddInt64(&r.bytesRead, int64(n))
	atomic.AddInt64(&r.messagesRead, 1)
	r.stats.SourceAddr = addr.String()
	r.stats.ConnectionTime = time.Since(r.stats.StartTime)

	// Create a reader from the UDP packet data
	packetReader := bytes.NewReader(buf[:n])

	// Decode the message
	msg, err := r.decoder.DecodeFrom(packetReader)
	if err != nil {
		// Log the error but don't return it so we can continue reading
		fmt.Fprintf(os.Stderr, "Warning: Error decoding ASTERIX message: %v\n", err)
		return nil, fmt.Errorf("decoding ASTERIX message: %w", err)
	}

	return msg, nil
}

// Close closes the underlying connection
func (r *udpAsterixReader) Close() error {
	return r.conn.Close()
}

// Protocol returns the transport protocol name
func (r *udpAsterixReader) Protocol() string {
	return "UDP"
}

// Stats returns reader statistics
func (r *udpAsterixReader) Stats() ReaderStats {
	// Create a copy with atomic loads
	stats := r.stats
	stats.BytesRead = atomic.LoadInt64(&r.bytesRead)
	stats.MessagesRead = atomic.LoadInt64(&r.messagesRead)
	stats.TransportErrors = int(atomic.LoadInt64(&r.transportErrors))
	stats.ConnectionTime = time.Since(r.stats.StartTime)
	return stats
}
