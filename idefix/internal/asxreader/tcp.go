// internal/asxreader/tcp.go
package asxreader

import (
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/davidkohl/gobelix/asterix"
)

// tcpAsterixReader implements AsterixReader for TCP connections
type tcpAsterixReader struct {
	conn      net.Conn
	listener  net.Listener
	decoder   *asterix.Decoder
	stats     ReaderStats
	lastError error

	// For atomic access to stats
	bytesRead       int64
	messagesRead    int64
	transportErrors int32 // Using int32 for atomic operations
}

// NewTCPAsterixReader creates a reader for TCP ASTERIX messages
func NewTCPAsterixReader(port int, decoder *asterix.Decoder) (AsterixReader, error) {
	// Check for nil decoder
	if decoder == nil {
		return nil, fmt.Errorf("decoder cannot be nil")
	}

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on TCP port %d: %w", port, err)
	}

	fmt.Fprintf(os.Stderr, "Waiting for TCP connection on port %d...\n", port)
	conn, err := listener.Accept()
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to accept TCP connection: %w", err)
	}

	// Set a default read deadline to prevent blocking indefinitely
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	return &tcpAsterixReader{
		conn:     conn,
		listener: listener,
		decoder:  decoder,
		stats:    NewReaderStats(),
	}, nil
}

// Next reads and decodes the next ASTERIX message from TCP
func (r *tcpAsterixReader) Next() (*asterix.DataBlock, error) {
	// Safety check for nil decoder or connection
	if r.decoder == nil {
		return nil, fmt.Errorf("nil decoder in TCP reader")
	}

	if r.conn == nil {
		return nil, fmt.Errorf("nil connection in TCP reader")
	}

	// Use the decoder directly with the connection
	msg, err := r.decoder.DecodeFrom(r.conn)
	if err != nil {
		r.lastError = err
		atomic.AddInt32(&r.transportErrors, 1)
		return nil, err
	}

	// Update statistics
	msgSize := msg.EstimateSize()
	atomic.AddInt64(&r.bytesRead, int64(msgSize))
	atomic.AddInt64(&r.messagesRead, 1)
	r.stats.SourceAddr = r.conn.RemoteAddr().String()
	r.stats.ConnectionTime = time.Since(r.stats.StartTime)

	return msg, nil
}

// Close closes the underlying connection and listener
func (r *tcpAsterixReader) Close() error {
	// Close the connection first
	connErr := r.conn.Close()

	// Then close the listener
	listenerErr := r.listener.Close()

	// Return the first error encountered
	if connErr != nil {
		return connErr
	}
	return listenerErr
}

// Protocol returns the transport protocol name
func (r *tcpAsterixReader) Protocol() string {
	return "TCP"
}

// Stats returns reader statistics
func (r *tcpAsterixReader) Stats() ReaderStats {
	// Create a copy to avoid race conditions
	return ReaderStats{
		BytesRead:       atomic.LoadInt64(&r.bytesRead),
		MessagesRead:    atomic.LoadInt64(&r.messagesRead),
		ConnectionTime:  time.Since(r.stats.StartTime),
		SourceAddr:      r.stats.SourceAddr,
		TransportErrors: int(atomic.LoadInt32(&r.transportErrors)),
		StartTime:       r.stats.StartTime,
	}
}

// SetReadDeadline sets a deadline for the next read from the TCP connection
func (r *tcpAsterixReader) SetReadDeadline(t time.Time) error {
	if r.conn == nil {
		return fmt.Errorf("nil TCP connection")
	}
	return r.conn.SetReadDeadline(t)
}
