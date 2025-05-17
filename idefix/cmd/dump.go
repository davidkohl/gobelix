// cmd/dump.go
package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davidkohl/gobelix/idefix/internal/asxreader"
	"github.com/davidkohl/gobelix/idefix/internal/decoder"
	"github.com/davidkohl/gobelix/idefix/internal/stats"
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
	dumpCmd.Flags().BoolVar(&dumpCat021, "dump021", false, "Dump ASTERIX category 021")
	dumpCmd.Flags().BoolVar(&dumpCat048, "dump048", false, "Dump ASTERIX category 048")
	dumpCmd.Flags().BoolVar(&dumpCat062, "dump062", false, "Dump ASTERIX category 062")
	dumpCmd.Flags().BoolVar(&dumpCat063, "dump063", false, "Dump ASTERIX category 063")

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

	// Create decoder with selected UAPs and optimizations
	asterixDecoder, err := decoder.CreateDecoder(decoder.Config{
		DumpAll:    dumpAll,
		DumpCat021: dumpCat021,
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
		return err
	}
	defer reader.Close()

	logger.Info("Listening for ASTERIX messages",
		"protocol", reader.Protocol(),
		"port", port)

	// Handle SIGINT for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create timeout channel if timeout is specified
	var timeoutCh <-chan time.Time
	if timeout > 0 {
		timeoutCh = time.After(time.Duration(timeout) * time.Second)
	}

	// Create stats ticker if stats are requested
	var statsTicker *time.Ticker
	if statsEvery > 0 {
		statsTicker = time.NewTicker(time.Duration(statsEvery) * time.Second)
		defer statsTicker.Stop()
	}

	// Track statistics
	messageStats := stats.NewMessageStats()

	// Start processing in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- processMessages(reader, out, logger, messageStats)
	}()

	// Wait for completion, interrupt, or timeout
	for {
		select {
		case err := <-done:
			if Verbose {
				messageStats.LogStats(logger, true)
			}
			return err
		case <-sigCh:
			logger.Info("Shutting down due to signal")
			messageStats.LogStats(logger, true)
			return nil
		case <-timeoutCh:
			logger.Info("Shutting down due to timeout", "timeout_seconds", timeout)
			messageStats.LogStats(logger, true)
			return nil
		case <-statsTicker.C:
			messageStats.LogStats(logger, false)
		}
	}
}

func processMessages(reader asxreader.AsterixReader, out *os.File, logger *slog.Logger, msgStats *stats.MessageStats) error {
	logger.Debug("Starting message processing loop")

	for {
		msg, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection closed")
				return nil // Connection closed normally
			}

			// Log the error but continue processing
			logger.Error("Error reading message", "error", err)
			continue
		}

		// Update statistics
		msgStats.IncrementCategory(msg.Category)

		// Print the message
		fmt.Fprintln(out, msg)

		logger.Debug("Processed message",
			"category", msg.Category.String(),
			"records", msg.GetRecordCount())
	}
}
