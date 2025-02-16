// example/main.go
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/dataitems/cat021"
)

func main() {
	// Create ASTERIX decoder with UAPs we want to support
	decoder, err := asterix.NewDecoder(
		cat021.NewUAP021(),
		// Add other UAPs as needed
	)
	if err != nil {
		fmt.Printf("Failed to create decoder: %v\n", err)
		return
	}

	// Connect to TCP server
	conn, err := net.Dial("tcp", "davidkohl.de:21000")
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// Handle graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create buffer for reading
	buf := make([]byte, 4096)
	remainder := make([]byte, 0)

	// Process loop
	for {
		select {
		case <-interrupt:
			fmt.Println("\nShutting down...")
			return
		default:
			// Read data from connection
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("Connection closed")
					return
				}
				fmt.Printf("Error reading from connection: %v\n", err)
				return
			}

			// Combine with remainder from previous read
			data := append(remainder, buf[:n]...)

			// Process complete messages
			offset := 0
			for offset < len(data) {
				// Need at least 3 bytes for category and length
				if len(data[offset:]) < 3 {
					remainder = data[offset:]
					break
				}

				// Get message length
				msgLen := uint16(data[offset+1])<<8 | uint16(data[offset+2])
				if len(data[offset:]) < int(msgLen) {
					remainder = data[offset:]
					break
				}

				// Decode message
				records, err := decoder.Decode(data[offset : offset+int(msgLen)])
				if err != nil {
					fmt.Printf("Failed to decode message: %v\n", err)
					offset += int(msgLen)
					continue
				}

				// Print decoded records
				for i, items := range records {
					fmt.Printf("\nRecord %d:\n", i+1)
					printItems(items)
				}

				offset += int(msgLen)
			}
		}
	}
}

func printItems(items map[string]asterix.DataItem) {
	for id, item := range items {
		switch v := item.(type) {
		default:
			fmt.Printf("  %s: %+v\n", id, v)
		}
	}
}
