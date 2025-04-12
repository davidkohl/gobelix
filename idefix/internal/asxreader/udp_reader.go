// udp_reader.go
package asxreader

import (
	"bytes"
	"fmt"
	"net"

	"github.com/davidkohl/gobelix/asterix"
)

// udpAsterixReader implements AsterixReader for UDP connections
type udpAsterixReader struct {
	conn      net.PacketConn
	buf       []byte
	decoder   *asterix.Decoder
	lastError error
}

// NewUDPAsterixReader creates a reader for UDP ASTERIX messages
func NewUDPAsterixReader(port int, decoder *asterix.Decoder) (AsterixReader, error) {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
	}

	return &udpAsterixReader{
		conn:    conn,
		buf:     make([]byte, 65536), // Max UDP packet size
		decoder: decoder,
	}, nil
}

// Next reads and decodes the next ASTERIX message from UDP
func (r *udpAsterixReader) Next() (*asterix.AsterixMessage, error) {
	n, _, err := r.conn.ReadFrom(r.buf)
	if err != nil {
		r.lastError = err
		return nil, fmt.Errorf("reading UDP packet: %w", err)
	}

	// Create a reader from the UDP packet data
	packetReader := bytes.NewReader(r.buf[:n])

	// Decode the message
	msg, err := r.decoder.Decode(packetReader)
	if err != nil {
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
