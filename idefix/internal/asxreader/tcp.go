// internal/asxreader/tcp.go
package net

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
}

// NewTCPAsterixReader creates a reader for TCP ASTERIX messages
func NewTCPAsterixReader(port int, decoder *asterix.Decoder) (AsterixReader, error) {
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

	return &tcpAsterixReader{
		conn:     conn,
		listener: listener,
		decoder:  decoder,
		stats:    NewReaderStats(),
	}, nil
}

// Next reads and decodes the next ASTERIX message from TCP
func (r *tcpAsterixReader) Next() (*asterix.DataBlock, error) {
	// Set read deadline if needed (e.g., to support timeout)
	// r.conn.SetReadDeadline(time.Now().Add(readTimeout))

	// Use the decoder directly with the connection
	msg, err := r.decoder.DecodeFrom(r.conn)
	if err != nil {
		r.lastError = err
		atomic.AddInt32((*int32)(&r.stats.TransportErrors), 1)
		return nil, err
	}

	// Update statistics
	msgSize := msg.EstimateSize()
	atomic.AddInt64(&r.stats.BytesRead, int64(msgSize))
	atomic.AddInt64(&r.stats.MessagesRead, 1)
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
		BytesRead:       atomic.LoadInt64(&r.stats.BytesRead),
		MessagesRead:    atomic.LoadInt64(&r.stats.MessagesRead),
		ConnectionTime:  r.stats.ConnectionTime,
		SourceAddr:      r.stats.SourceAddr,
		TransportErrors: int(atomic.LoadInt32((*int32)(&r.stats.TransportErrors))),
		StartTime:       r.stats.StartTime,
	}
}
