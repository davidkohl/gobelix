// cmd/dump.go
package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat021"
	"github.com/davidkohl/gobelix/cat/cat062"

	//"github.com/davidkohl/gobelix/cat/cat063"
	"github.com/spf13/cobra"
)

var (
	portFlag   string
	outputFile string
	dumpAll    bool
	dumpCat021 bool
	dumpCat062 bool
	dumpCat063 bool
)

func init() {
	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump ASTERIX messages from network traffic",
		Long: `Listen on a specified port and dump decoded ASTERIX messages to stdout or a file.
Example: idefix dump -p 2000/udp --dump021`,
		RunE: runDump,
	}

	// Add port flag
	dumpCmd.Flags().StringVarP(&portFlag, "port", "p", "", "Port to listen on with protocol (e.g., 2000/udp)")
	dumpCmd.MarkFlagRequired("port")

	// Add output flag
	dumpCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	// Add category flags
	dumpCmd.Flags().BoolVar(&dumpAll, "dumpAll", false, "Dump all ASTERIX categories")
	dumpCmd.Flags().BoolVar(&dumpCat021, "dump021", false, "Dump ASTERIX category 021")
	dumpCmd.Flags().BoolVar(&dumpCat062, "dump062", false, "Dump ASTERIX category 062")
	dumpCmd.Flags().BoolVar(&dumpCat063, "dump063", false, "Dump ASTERIX category 063")

	rootCmd.AddCommand(dumpCmd)
}

func runDump(cmd *cobra.Command, args []string) error {
	// Parse port and protocol
	parts := strings.Split(portFlag, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid port format, use PORT/PROTOCOL, e.g., 2000/udp")
	}

	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid port number: %w", err)
	}

	protocol := strings.ToLower(parts[1])
	if protocol != "udp" && protocol != "tcp" {
		return fmt.Errorf("protocol must be either 'udp' or 'tcp'")
	}

	// Setup output
	var out *os.File
	if outputFile == "" {
		out = os.Stdout
	} else {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
	}

	// Create decoder with selected UAPs
	decoder, err := createDecoder()
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	// Setup connection
	var conn io.ReadCloser
	addr := fmt.Sprintf(":%d", port)

	// Handle SIGINT for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Setup connection based on protocol
	switch protocol {
	case "udp":
		udpConn, err := net.ListenPacket("udp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on UDP port %s: %w", parts[0], err)
		}
		defer udpConn.Close()
		conn = newUDPReader(udpConn)

	case "tcp":
		tcpListener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on TCP port %s: %w", parts[0], err)
		}
		defer tcpListener.Close()

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Fprintf(os.Stderr, "Waiting for TCP connection on port %s...\n", parts[0])
		}

		// Accept the first connection
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept TCP connection: %w", err)
		}
		conn = tcpConn
	}

	defer conn.Close()

	// Create reader
	reader := asterix.NewReader(conn, decoder)
	defer reader.Close()

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Fprintf(os.Stderr, "Listening for ASTERIX messages on %s port %s...\n",
			protocol, parts[0])
	}

	// Start processing in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- processMessages(reader, out)
	}()

	// Wait for either completion or interrupt
	select {
	case err := <-done:
		return err
	case <-sigCh:
		if verbose {
			fmt.Fprintf(os.Stderr, "\nShutting down...\n")
		}
		return nil
	}
}

func processMessages(reader *asterix.Reader, out *os.File) error {
	for {
		msg, err := reader.ReadMessage()
		if err != nil {
			if err == io.EOF {
				return nil // Connection closed normally
			}
			// Log the error but continue processing
			fmt.Fprintf(os.Stderr, "Error reading message: %v\n", err)
			continue
		}

		// Print the message
		fmt.Fprintln(out, msg)

		// If we want more detailed record output, we could do:
		// for i := 0; i < msg.GetRecordCount(); i++ {
		//     // Print each record - this assumes we've added a way to get records
		// }
	}
}

func createDecoder() (*asterix.Decoder, error) {
	var uaps []asterix.UAP

	if dumpAll || dumpCat021 {
		uap021, err := cat021.NewUAP("2.6")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat021 UAP: %w", err)
		}
		uaps = append(uaps, uap021)
	}

	if dumpAll || dumpCat062 {
		uap062, err := cat062.NewUAP("1.17")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat062 UAP: %w", err)
		}
		uaps = append(uaps, uap062)
	}

	if len(uaps) == 0 {
		return nil, fmt.Errorf("no categories selected, use --dumpAll or specify categories")
	}

	return asterix.NewDecoder(uaps...)
}

// UDPReader adapts a net.PacketConn to be an io.ReadCloser
type UDPReader struct {
	conn net.PacketConn
	buf  []byte
}

func newUDPReader(conn net.PacketConn) *UDPReader {
	return &UDPReader{
		conn: conn,
		buf:  make([]byte, 65536), // Max UDP packet size
	}
}

func (r *UDPReader) Read(p []byte) (n int, err error) {
	n, _, err = r.conn.ReadFrom(r.buf)
	if err != nil {
		return 0, err
	}

	if n > len(p) {
		// Packet is larger than the provided buffer
		return copy(p, r.buf[:len(p)]), io.ErrShortBuffer
	}

	return copy(p, r.buf[:n]), nil
}

func (r *UDPReader) Close() error {
	return r.conn.Close()
}
