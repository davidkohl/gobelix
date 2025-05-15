// examples/simple_example.go
package main

import (
	"fmt"
	"log"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat021"
	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func main() {
	fmt.Println("ASTERIX Category 021 Simple Example")
	fmt.Println("----------------------------------")

	// Get the Category 021 UAP
	uap, err := cat021.NewUAP(cat021.Version26)
	if err != nil {
		log.Fatalf("Failed to create UAP: %v", err)
	}

	// Create a data block for encoding
	dataBlock, err := asterix.NewDataBlock(asterix.Cat021, uap)
	if err != nil {
		log.Fatalf("Failed to create data block: %v", err)
	}

	// Create a record
	record, err := asterix.NewRecord(asterix.Cat021, uap)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}

	// Add required data items based on the UAP
	// 1. Data Source Identifier (mandatory)
	dsi := &common.DataSourceIdentifier{
		SAC: 25,  // System Area Code
		SIC: 100, // System Identification Code
	}
	if err := record.SetDataItem("I021/010", dsi); err != nil {
		log.Fatalf("Failed to set Data Source Identifier: %v", err)
	}

	// 2. Target Report Descriptor (mandatory)
	trd := &v26.TargetReportDescriptor{
		ATP: 0, // 24-bit ICAO address
		ARC: 0, // 25ft resolution
	}
	if err := record.SetDataItem("I021/040", trd); err != nil {
		log.Fatalf("Failed to set Target Report Descriptor: %v", err)
	}

	// 3. Target Address (mandatory)
	ta := &v26.TargetAddress{
		Address: 0xABCDEF, // ICAO address
	}
	if err := record.SetDataItem("I021/080", ta); err != nil {
		log.Fatalf("Failed to set Target Address: %v", err)
	}

	// Add the record to the data block
	if err := dataBlock.AddRecord(record); err != nil {
		log.Fatalf("Failed to add record: %v", err)
	}

	// Encode the data block
	fmt.Println("Encoding ASTERIX message...")
	data, err := dataBlock.Encode()
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	fmt.Printf("Encoded data (%d bytes): %X\n", len(data), data)

	// Decode the data
	fmt.Println("\nDecoding ASTERIX message...")
	decodedBlock, err := asterix.NewDataBlock(asterix.Cat021, uap)
	if err != nil {
		log.Fatalf("Failed to create decode block: %v", err)
	}

	err = decodedBlock.Decode(data)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	// Extract and print the decoded record
	if decodedBlock.RecordCount() == 0 {
		log.Fatal("No records found in decoded data block")
	}

	fmt.Println("Successfully decoded message with the following items:")
	decodedRecord := decodedBlock.Records()[0]

	// Print all decoded items
	for _, field := range uap.Fields() {
		if item, exists := decodedRecord.GetDataItem(field.DataItem); exists {
			fmt.Printf("  %s (%s): %v\n", field.DataItem, field.Description, item)
		}
	}

	fmt.Println("\nExample completed successfully!")
}
