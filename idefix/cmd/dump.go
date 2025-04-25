// dump.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat021"
	"github.com/davidkohl/gobelix/cat/cat048"
	"github.com/davidkohl/gobelix/cat/cat062"
	"github.com/davidkohl/gobelix/cat/cat063"
	"github.com/davidkohl/gobelix/idefix/internal/asxreader"
	"github.com/spf13/cobra"
)

var (
	portFlag   string
	outputFile string
	dumpAll    bool
	dumpCat021 bool
	dumpCat048 bool
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
	dumpCmd.Flags().BoolVar(&dumpCat048, "dump048", false, "Dump ASTERIX category 048")
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

	// Create the ASTERIX reader
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Fprintf(os.Stderr, "Creating %s reader on port %s...\n", protocol, parts[0])
	}

	reader, err := asxreader.NewAsterixReader(protocol, port, decoder)
	if err != nil {
		return err
	}
	defer reader.Close()

	if verbose {
		fmt.Fprintf(os.Stderr, "Listening for ASTERIX messages on %s port %s...\n",
			reader.Protocol(), parts[0])
	}

	// Handle SIGINT for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start processing in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- processMessages(reader, out, verbose)
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

func processMessages(reader asxreader.AsterixReader, out *os.File, verbose bool) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "Starting message processing loop...\n")
	}

	for {
		msg, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				if verbose {
					fmt.Fprintf(os.Stderr, "Connection closed\n")
				}
				return nil // Connection closed normally
			}

			// Log the error but continue processing
			fmt.Fprintf(os.Stderr, "Error reading message: %v\n", err)
			return err
		}

		// Print the message
		fmt.Fprintln(out, msg)

		if verbose {
			fmt.Fprintf(os.Stderr, "Processed message with %d records\n", msg.GetRecordCount())
		}
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

	if dumpAll || dumpCat048 {
		uap048, err := cat048.NewUAP("1.32")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat062 UAP: %w", err)
		}
		uaps = append(uaps, uap048)
	}

	if dumpAll || dumpCat062 {
		uap062, err := cat062.NewUAP("1.17")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat062 UAP: %w", err)
		}
		uaps = append(uaps, uap062)
	}

	if dumpAll || dumpCat063 {
		uap063, err := cat063.NewUAP("1.6")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat062 UAP: %w", err)
		}
		uaps = append(uaps, uap063)
	}

	if len(uaps) == 0 {
		return nil, fmt.Errorf("no categories selected, use --dumpAll or specify categories")
	}

	return asterix.NewDecoder(uaps...)
}
