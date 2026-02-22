# ASTERIX Category 021 - ADS-B Reports

This package implements ASTERIX Category 021 (ADS-B Reports) according to the EUROCONTROL specification.

## Purpose

Category 021 is used to transmit ADS-B surveillance information derived from 1090 MHz Extended Squitter, UAT or VDL Mode 4 ADS-B ground stations. It provides detailed aircraft state vector, status, and intent information from ADS-B equipped aircraft, including:

- Aircraft identification and addressing
- Position in WGS-84 coordinates (high and low resolution)
- Velocity (airborne and surface)
- Aircraft derived data (selected altitude, heading, speed, etc.)
- Target status and trajectory intent
- Mode S MB data
- ACAS resolution advisory reports

## Implementation

The implementation follows the ASTERIX Category 021 specification version 2.6 and includes:

- Complete User Application Profile (UAP) definition
- All required and optional data items
- Support for Reserved Expansion Field (REF) for future extensions
- Support for Special Purpose Field (SPF) for non-standard information
- Comprehensive encoding/decoding of all defined elements

## Data Items

The following data items are implemented:

| FRN | Data Item        | Description                           | Format    | Length | Mandatory |
|-----|------------------|---------------------------------------|-----------|--------|-----------|
| 1   | I021/010         | Data Source Identifier                | Fixed     | 2      | Yes       |
| 2   | I021/040         | Target Report Descriptor              | Extended  | 1+     | Yes       |
| 3   | I021/030         | Time of Day                           | Fixed     | 3      | No        |
| 4   | I021/130         | Position in WGS-84 Coordinates        | Fixed     | 6      | No        |
| 5   | I021/080         | Target Address                        | Fixed     | 3      | Yes       |
| 6   | I021/090         | Figure of Merit (Quality Indicators)  | Extended  | 1+     | No        |
| 7   | I021/210         | MOPS Version                          | Fixed     | 1      | No        |
| 8   | I021/070         | Mode 3/A Code                         | Fixed     | 2      | No        |
| 9   | I021/230         | Roll Angle                            | Fixed     | 2      | No        |
| 10  | I021/145         | Flight Level                          | Fixed     | 2      | No        |
| 11  | I021/150         | Air Speed                             | Fixed     | 2      | No        |
| 12  | I021/151         | True Air Speed                        | Fixed     | 2      | No        |
| 13  | I021/152         | Magnetic Heading                      | Fixed     | 2      | No        |
| 14  | I021/155         | Barometric Vertical Rate              | Fixed     | 2      | No        |
| 15  | I021/157         | Geometric Vertical Rate               | Fixed     | 2      | No        |
| 16  | I021/160         | Airborne Ground Vector                | Fixed     | 4      | No        |
| 17  | I021/165         | Track Angle Rate                      | Fixed     | 2      | No        |
| 18  | I021/170         | Target Identification                 | Fixed     | 6      | No        |
| 19  | I021/095         | Velocity Accuracy                     | Fixed     | 1      | No        |
| 20  | I021/032         | Time of Day Accuracy                  | Fixed     | 1      | No        |
| 21  | I021/200         | Target Status                         | Fixed     | 1      | No        |
| 22  | I021/020         | Emitter Category                      | Fixed     | 1      | No        |
| 23  | I021/220         | Met Information                       | Compound  | 1+     | No        |
| 24  | I021/146         | Selected Altitude                     | Fixed     | 2      | No        |
| 25  | I021/148         | Final State Selected Altitude         | Fixed     | 2      | No        |
| 26  | I021/110         | Trajectory Intent                     | Compound  | 1+     | No        |
| 27  | I021/016         | Service Management                    | Fixed     | 1      | No        |
| 28  | I021/008         | Aircraft Operational Status           | Fixed     | 1      | No        |
| 29  | I021/271         | Surface Capabilities and Characteristics | Extended | 1+   | No        |
| 30  | I021/132         | Message Amplitude                     | Fixed     | 1      | No        |
| 31  | I021/250         | Mode S MB Data                        | Repetitive| 8+     | No        |
| 32  | I021/260         | ACAS Resolution Advisory Report       | Fixed     | 7      | No        |
| 33  | I021/400         | Receiver ID                           | Fixed     | 1      | No        |
| 34  | I021/295         | Data Ages                             | Compound  | 1+     | No        |
| 35  | RE021            | Reserved Expansion Field              | Repetitive| 1+     | No        |
| 36  | SP021            | Special Purpose Field                 | Repetitive| 1+     | No        |
| 37  | I021/131         | Position in WGS-84 Coordinates (High Resolution) | Fixed | 8 | No    |
| 38  | I021/072         | Time of Applicability for Velocity    | Fixed     | 3      | No        |
| 39  | I021/073         | Time of Message Reception Position    | Fixed     | 3      | No        |
| 40  | I021/074         | Time of Message Reception Position High Precision | Fixed | 4 | No   |
| 41  | I021/075         | Time of Message Reception Velocity    | Fixed     | 3      | No        |
| 42  | I021/076         | Time of Message Reception Velocity High Precision | Fixed | 4 | No   |
| 43  | I021/077         | Time of Report Transmission           | Fixed     | 3      | No        |
| 44  | I021/140         | Geometric Height                      | Fixed     | 2      | No        |
| 45  | I021/290         | System Status                         | Fixed     | 1      | No        |
| 46  | I021/071         | Time of Applicability for Position    | Fixed     | 3      | No        |

## Usage

Here's a basic example showing how to create and encode a Category 021 message:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat021"
    v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
    common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func main() {
    // Create a UAP
    uap, err := cat021.NewUAP(cat021.Version26)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        os.Exit(1)
    }
    
    // Create a new data block
    db, err := asterix.NewDataBlock(asterix.Cat021, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create data block: %v\n", err)
        os.Exit(1)
    }
    
    // Create a record
    record, err := asterix.NewRecord(asterix.Cat021, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create record: %v\n", err)
        os.Exit(1)
    }
    
    // Set mandatory items
    record.SetDataItem("I021/010", &common.DataSourceIdentifier{SAC: 25, SIC: 10})
    
    // Create target report descriptor
    trd := &v26.TargetReportDescriptor{
        ATP: 1, // 1090 ES
        ARC: 0, // 25 ft resolution
        RC:  false,
        RAB: false,
    }
    record.SetDataItem("I021/040", trd)
    
    // Set target address (24-bit ICAO address)
    record.SetDataItem("I021/080", &v26.TargetAddress{Address: 0xABC123})
    
    // Add position (latitude/longitude)
    position := &common.Position{
        Latitude:  51.5074, // London
        Longitude: -0.1278,
    }
    record.SetDataItem("I021/130", position)
    
    // Add flight level
    fl := &common.FlightLevel{
        Value: 350.0, // FL350
    }
    record.SetDataItem("I021/145", fl)
    
    // Add identification (callsign)
    callsign := &v26.TargetIdentification{
        Ident: "BAW123",
    }
    record.SetDataItem("I021/170", callsign)
    
    // Add the record to the data block
    err = db.AddRecord(record)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to add record: %v\n", err)
        os.Exit(1)
    }
    
    // Encode the data block
    encoded, err := db.Encode()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to encode data block: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Successfully encoded Category 021 message (%d bytes)\n", len(encoded))
}
```

## Decoding Example

Here's an example of decoding a received Category 021 message:

```go
func decodeMessage(data []byte) {
    // Create a UAP
    uap, err := cat021.NewUAP(cat021.Version26)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        return
    }
    
    // Create an empty data block
    db, err := asterix.NewDataBlock(asterix.Cat021, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create data block: %v\n", err)
        return
    }
    
    // Decode the message
    err = db.Decode(data)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to decode data: %v\n", err)
        return
    }
    
    // Access the decoded records
    records := db.Records()
    fmt.Printf("Decoded %d records\n", len(records))
    
    for i, record := range records {
        fmt.Printf("Record #%d:\n", i+1)
        
        // Access data source identifier
        if dataSource, _, exists := record.GetDataItem("I021/010"); exists {
            dsi := dataSource.(*common.DataSourceIdentifier)
            fmt.Printf("  Data Source: SAC=%d, SIC=%d\n", dsi.SAC, dsi.SIC)
        }
        
        // Access target address
        if targetAddr, _, exists := record.GetDataItem("I021/080"); exists {
            ta := targetAddr.(*v26.TargetAddress)
            fmt.Printf("  Target Address: %06X\n", ta.Address)
        }
        
        // Access position
        if position, _, exists := record.GetDataItem("I021/130"); exists {
            pos := position.(*common.Position)
            fmt.Printf("  Position: Lat=%.6f°, Lon=%.6f°\n", pos.Latitude, pos.Longitude)
        }
        
        // Access flight level
        if flightLevel, _, exists := record.GetDataItem("I021/145"); exists {
            fl := flightLevel.(*common.FlightLevel)
            fmt.Printf("  Flight Level: FL%.0f\n", fl.Value)
        }
        
        // Access target identification
        if targetId, _, exists := record.GetDataItem("I021/170"); exists {
            ti := targetId.(*v26.TargetIdentification)
            fmt.Printf("  Callsign: %s\n", ti.Ident)
        }
    }
}
```

## Notes

- ADS-B positions are generally provided with higher accuracy than traditional radar
- The implementation handles both airborne and surface position reports
- Quality indicators provide information about position, velocity, and barometric altitude accuracy
- All numeric values use the appropriate scaling factors as defined in the specification
- Full support for extended items with the Field Extension (FX) mechanism