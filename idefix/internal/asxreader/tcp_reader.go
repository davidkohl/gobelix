// tcp_reader.go
package asxreader

import (
	"fmt"
	"net"
	"os"

	"github.com/davidkohl/gobelix/asterix"
)

// tcpAsterixReader implements AsterixReader for TCP connections
type tcpAsterixReader struct {
	conn      net.Conn
	listener  net.Listener
	decoder   *asterix.Decoder
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
	}, nil
}

// Next reads and decodes the next ASTERIX message from TCP
func (r *tcpAsterixReader) Next() (*asterix.AsterixMessage, error) {
	msg, err := r.decoder.Decode(r.conn)
	if err != nil {
		r.lastError = err
		return nil, err
	}

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
