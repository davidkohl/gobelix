# ASTERIX Category 063 - Sensor Status Reports

This package implements ASTERIX Category 063 (Sensor Status Reports) according to the EUROCONTROL specification version 1.6.

## Purpose

Category 063 is used to transmit the status and configuration of surveillance sensors and system status information from a Surveillance Data Processing System (SDPS) to users. It provides information such as:

- Sensor operating status (operational, degraded, initialization, etc.)
- Configuration details
- Time stamping bias
- Range, azimuth, and elevation bias/gain values
- Other sensor-specific parameters

## Implementation

The implementation follows the ASTERIX Category 063 specification version 1.6 and includes:

- Complete User Application Profile (UAP) definition
- All mandatory and optional data items
- Support for Reserved Expansion Field (REF) for sensor status information specific to an SDPS
- Support for Special Purpose Field (SPF) for non-standard information

## Data Items

The following data items are implemented:

| FRN | Data Item        | Description                        | Format  | Length | Mandatory |
|-----|------------------|------------------------------------|---------|--------|-----------|
| 1   | I063/010         | Data Source Identifier             | Fixed   | 2      | Yes       |
| 2   | I063/015         | Service Identification             | Fixed   | 1      | No        |
| 3   | I063/030         | Time of Message                    | Fixed   | 3      | Yes       |
| 4   | I063/050         | Sensor Identifier                  | Fixed   | 2      | Yes       |
| 5   | I063/060         | Sensor Configuration and Status    | Extended| 1+1    | No        |
| 6   | I063/070         | Time Stamping Bias                 | Fixed   | 2      | No        |
| 7   | I063/080         | SSR/Mode S Range Gain and Bias     | Fixed   | 4      | No        |
| 8   | I063/081         | SSR/Mode S Azimuth Bias            | Fixed   | 2      | No        |
| 9   | I063/090         | PSR Range Gain and Bias            | Fixed   | 4      | No        |
| 10  | I063/091         | PSR Azimuth Bias                   | Fixed   | 2      | No        |
| 11  | I063/092         | PSR Elevation Bias                 | Fixed   | 2      | No        |
| 12  | -                | Spare                              | -       | -      | No        |
| 13  | RE063            | Reserved Expansion Field           | Repetitive | 1+  | No        |
| 14  | SP063            | Special Purpose Field              | Repetitive | 1+  | No        |

## Usage

Here's a basic example showing how to create and encode a Category 063 message:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat063"
    v16 "github.com/davidkohl/gobelix/cat/cat063/dataitems/v16"
    "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func main() {
    // Create a UAP
    uap, err := cat063.NewUAP(cat063.Version16)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        os.Exit(1)
    }
    
    // Create a new data block
    db, err := asterix.NewDataBlock(asterix.Cat063, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create data block: %v\n", err)
        os.Exit(1)
    }
    
    // Create a record
    record, err := asterix.NewRecord(asterix.Cat063, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create record: %v\n", err)
        os.Exit(1)
    }
    
    // Set mandatory items
    record.SetDataItem("I063/010", &dataitems.DataSourceIdentifier{SAC: 25, SIC: 10})
    record.SetDataItem("I063/030", &v16.TimeOfMessage{Time: 36000.5}) // 10:00:00.5
    record.SetDataItem("I063/050", &v16.SensorIdentifier{SAC: 42, SIC: 123})
    
    // Add optional items
    record.SetDataItem("I063/060", &v16.SensorConfigurationAndStatus{
        CON:           v16.StatusOperational,
        PSR:           false,
        SSR:           false,
        MDS:           true,  // Mode S NOGO
        ADS:           false,
        MLT:           false,
        HasFirstExtent: true,
        MSC:           true,  // Monitoring system disconnected
    })
    
    record.SetDataItem("I063/070", &v16.TimeStampingBias{Bias: -120}) // -120ms
    
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
    
    // Write to a file or network
    // os.WriteFile("cat063_message.bin", encoded, 0644)
    
    fmt.Printf("Successfully encoded Category 063 message (%d bytes)\n", len(encoded))
}
```

## Decoding Example

Here's an example of decoding a received Category 063 message:

```go
func decodeMessage(data []byte) {
    // Create a UAP
    uap, err := cat063.NewUAP(cat063.Version16)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        return
    }
    
    // Create an empty data block
    db, err := asterix.NewDataBlock(asterix.Cat063, uap)
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
        
        // Access data source identifier (mandatory)
        if dataSource, _, exists := record.GetDataItem("I063/010"); exists {
            dsi := dataSource.(*dataitems.DataSourceIdentifier)
            fmt.Printf("  Data Source: SAC=%d, SIC=%d\n", dsi.SAC, dsi.SIC)
        }
        
        // Access time of message (mandatory)
        if timeOfMsg, _, exists := record.GetDataItem("I063/030"); exists {
            tom := timeOfMsg.(*v16.TimeOfMessage)
            fmt.Printf("  Time: %s\n", tom.String())
        }
        
        // Access sensor identifier (mandatory)
        if sensorID, _, exists := record.GetDataItem("I063/050"); exists {
            si := sensorID.(*v16.SensorIdentifier)
            fmt.Printf("  Sensor: SAC=%d, SIC=%d\n", si.SAC, si.SIC)
        }
        
        // Access sensor status (optional)
        if sensorStatus, _, exists := record.GetDataItem("I063/060"); exists {
            ss := sensorStatus.(*v16.SensorConfigurationAndStatus)
            fmt.Printf("  Status: %s\n", ss.String())
        }
        
        // Access time stamping bias (optional)
        if timeStampingBias, _, exists := record.GetDataItem("I063/070"); exists {
            tsb := timeStampingBias.(*v16.TimeStampingBias)
            fmt.Printf("  Time Bias: %s\n", tsb.String())
        }
    }
}
```

## Notes

- All numeric values using two's complement representation (for bias values) are properly handled
- Proper validation is implemented for all data items
- Comprehensive testing ensures the implementation complies with the specification
- The implementation handles both encoding and decoding of messages

## Sensor Types

The Sensor Configuration and Status data item can indicate the status of various surveillance sensors:

- PSR: Primary Surveillance Radar
- SSR: Secondary Surveillance Radar
- MDS: Mode S
- ADS: Automatic Dependent Surveillance (ADS-B)
- MLT: Multilateration

## Connection Status

The following connection statuses are defined:

- Operational: Sensor operating normally
- Degraded: Sensor operating with reduced capability
- Initialization: Sensor initializing
- Not Connected: Sensor not connected to SDPS