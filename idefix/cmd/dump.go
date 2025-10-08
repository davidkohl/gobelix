// cmd/dump.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/idefix/internal/asxreader"
	"github.com/davidkohl/gobelix/idefix/internal/decoder"
	"github.com/davidkohl/gobelix/idefix/internal/stats"
	"github.com/spf13/cobra"
)

var (
	portFlag   string
	outputFile string
	dumpAll    bool
	dumpCat001 bool
	dumpCat002 bool
	dumpCat020 bool
	dumpCat021 bool
	dumpCat034 bool
	dumpCat048 bool
	dumpCat062 bool
	dumpCat063 bool
	timeout    int
	statsEvery int
)

func init() {
	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump ASTERIX messages from network traffic",
		Long: `Listen on a specified port and dump decoded ASTERIX messages to stdout or a file.
Example: idefix dump -p 2000/udp --dump021

This command listens for ASTERIX messages on the specified port and protocol,
decodes them according to the selected categories, and outputs the decoded
information in a human-readable format.`,
		Example: `  # Dump Category 021 messages from UDP port 2000
  idefix dump -p 2000/udp --dump021
  
  # Dump all categories from TCP port 8600 and save to file
  idefix dump -p 8600/tcp --dumpAll -o asterix_data.txt
  
  # Dump Categories 021 and 062 from UDP port 10001
  idefix dump -p 10001/udp --dump021 --dump062`,
		RunE: runDump,
	}

	// Add port flag
	dumpCmd.Flags().StringVarP(&portFlag, "port", "p", "", "Port to listen on with protocol (e.g., 2000/udp)")
	dumpCmd.MarkFlagRequired("port")

	// Add output flag
	dumpCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	// Add category flags
	dumpCmd.Flags().BoolVar(&dumpAll, "dumpAll", false, "Dump all ASTERIX categories")
	dumpCmd.Flags().BoolVar(&dumpCat001, "dump001", false, "Dump ASTERIX category 001 (Monoradar Track)")
	dumpCmd.Flags().BoolVar(&dumpCat002, "dump002", false, "Dump ASTERIX category 002 (Monoradar Plot)")
	dumpCmd.Flags().BoolVar(&dumpCat020, "dump020", false, "Dump ASTERIX category 020 (Multilateration)")
	dumpCmd.Flags().BoolVar(&dumpCat021, "dump021", false, "Dump ASTERIX category 021 (ADS-B)")
	dumpCmd.Flags().BoolVar(&dumpCat034, "dump034", false, "Dump ASTERIX category 034 (Service Messages)")
	dumpCmd.Flags().BoolVar(&dumpCat048, "dump048", false, "Dump ASTERIX category 048 (Monoradar Target)")
	dumpCmd.Flags().BoolVar(&dumpCat062, "dump062", false, "Dump ASTERIX category 062 (Track)")
	dumpCmd.Flags().BoolVar(&dumpCat063, "dump063", false, "Dump ASTERIX category 063 (Sensor Status)")

	// Add additional flags
	dumpCmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout in seconds (0 = no timeout)")
	dumpCmd.Flags().IntVar(&statsEvery, "stats", 0, "Print stats every N seconds (0 = no stats)")

	rootCmd.AddCommand(dumpCmd)
}

func runDump(cmd *cobra.Command, args []string) error {
	// Configure logging
	logger := ConfigureLogger(Verbose, JsonLogs)

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
	logger.Info("Creating ASTERIX decoder")
	asterixDecoder, err := decoder.CreateDecoder(decoder.Config{
		DumpAll:    dumpAll,
		DumpCat001: dumpCat001,
		DumpCat002: dumpCat002,
		DumpCat020: dumpCat020,
		DumpCat021: dumpCat021,
		DumpCat034: dumpCat034,
		DumpCat048: dumpCat048,
		DumpCat062: dumpCat062,
		DumpCat063: dumpCat063,
	})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	// Create the ASTERIX reader
	logger.Info("Creating ASTERIX reader",
		"protocol", protocol,
		"port", port)

	reader, err := asxreader.NewAsterixReader(protocol, port, asterixDecoder)
	if err != nil {
		return fmt.Errorf("failed to create ASTERIX reader: %w", err)
	}
	defer reader.Close()

	logger.Info("Listening for ASTERIX messages",
		"protocol", reader.Protocol(),
		"port", port)

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Setup timeout if specified
	if timeout > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(timeout) * time.Second):
				logger.Info("Timeout reached, initiating shutdown", "timeout_seconds", timeout)
				cancel()
			case <-ctx.Done():
				// Context was canceled elsewhere, nothing to do
				return
			}
		}()
	}

	// Track statistics
	messageStats := stats.NewMessageStats()

	// Setup stats reporting if enabled
	if statsEvery > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(statsEvery) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					messageStats.LogStats(logger, false)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Start the message processing in a goroutine
	processDone := make(chan error, 1)
	go func() {
		processDone <- processMessages(ctx, reader, out, logger, messageStats, dumpAll)
	}()

	// Wait for signal or processing completion
	var result error
	select {
	case <-sigCh:
		logger.Info("Received shutdown signal, terminating")
		cancel()
		// Wait for processing to finish with a timeout
		select {
		case err := <-processDone:
			result = err
		case <-time.After(2 * time.Second):
			logger.Info("Forced shutdown after timeout")
		}
	case err := <-processDone:
		logger.Info("Message processing completed")
		result = err
	}

	// Print final statistics
	messageStats.LogStats(logger, true)
	return result
}

func processMessages(
	ctx context.Context,
	reader asxreader.AsterixReader,
	out *os.File,
	logger *slog.Logger,
	msgStats *stats.MessageStats,
	dumpAll bool,
) error {
	logger.Debug("Starting message processing loop")

	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			logger.Info("Message processing canceled")
			return nil
		default:
			// Continue processing
		}

		// Set a short read timeout to prevent blocking indefinitely on Next()
		// This is especially important for UDP
		if setDeadliner, ok := reader.(asxreader.DeadlineSetter); ok {
			setDeadliner.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		}

		msg, err := reader.Next()
		if err != nil {
			// Check for standard errors
			if err == io.EOF {
				logger.Info("Connection closed")
				return nil
			}

			// For timeout errors, just continue the loop to check for cancellation
			if isTimeoutError(err) {
				continue
			}

			// Suppress unknown category errors unless dumpAll is enabled or verbose logging
			if shouldSuppressError(err, dumpAll, Verbose) {
				continue
			}

			// Log other errors but keep running
			logger.Error("Error reading message", "error", err)
			continue
		}

		// Update statistics
		msgStats.IncrementCategory(msg.Category())

		// Print the message using its String() method
		fmt.Fprintln(out, msg.String())

		logger.Debug("Processed message",
			"category", msg.Category().String(),
			"records", msg.RecordCount())
	}
}

// shouldSuppressError determines if an error should be suppressed based on context
func shouldSuppressError(err error, dumpAll bool, verbose bool) bool {
	if err == nil {
		return false
	}

	// Never suppress errors if dumpAll is enabled or verbose logging is on
	if dumpAll || verbose {
		return false
	}

	// Suppress unknown category errors when specific categories are selected
	if errors.Is(err, asterix.ErrUAPNotDefined) || errors.Is(err, asterix.ErrUnknownCategory) {
		return true
	}

	return false
}

// isTimeoutError checks if an error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// Check for standard net timeout errors
	if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
		return true
	}

	// Check based on error string (less reliable but catches more cases)
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "deadline exceeded")
}
