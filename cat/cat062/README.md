# ASTERIX Category 062 - System Track Data

This package implements ASTERIX Category 062 (System Track Data) according to the EUROCONTROL specification.

## Purpose

Category 062 is used to transmit track information from surveillance data processing systems. It carries system track data that may be derived from various surveillance sources, including:

- Radar (PSR/SSR)
- Mode S
- ADS-B
- Multilateration
- Track fusion from multiple sensors

The category provides a standardized format for distributing system track information containing position, kinematics, identification, flight plan related data, and track quality indicators.

## Implementation

The implementation follows the ASTERIX Category 062 specification with support for versions:

- Version 1.17
- Version 1.20

Key features include:

- Complete User Application Profile (UAP) definition for each version
- All data items as specified in the standard
- Support for compound and extended data items
- Comprehensive validation of data fields
- Proper handling of calculated values (position, velocity, acceleration)

## Data Items

The following data items are implemented:

| FRN | Data Item        | Description                        | Format    | Length | Mandatory |
|-----|------------------|------------------------------------|-----------|--------|-----------|
| 1   | I062/010         | Data Source Identifier             | Fixed     | 2      | Yes       |
| 2   | -                | Spare                              | -         | -      | No        |
| 3   | I062/015         | Service Identification             | Fixed     | 1      | No        |
| 4   | I062/070         | Time of Track Information          | Fixed     | 3      | Yes       |
| 5   | I062/105         | Calculated Position in WGS-84      | Fixed     | 8      | No        |
| 6   | I062/100         | Calculated Track Position (Cartesian) | Fixed  | 6      | No        |
| 7   | I062/185         | Calculated Track Velocity          | Fixed     | 4      | No        |
| 8   | I062/210         | Calculated Acceleration            | Fixed     | 2      | No        |
| 9   | I062/060         | Track Mode 3/A Code                | Fixed     | 2      | No        |
| 10  | I062/245         | Target Identification              | Fixed     | 7      | No        |
| 11  | I062/380         | Aircraft Derived Data              | Compound  | 1+     | No        |
| 12  | I062/040         | Track Number                       | Fixed     | 2      | Yes       |
| 13  | I062/080         | Track Status                       | Extended  | 1+     | Yes       |
| 14  | I062/290         | System Track Update Ages           | Compound  | 1+     | No        |
| 15  | I062/200         | Mode of Movement                   | Fixed     | 1      | No        |
| 16  | I062/295         | Track Data Ages                    | Compound  | 1+     | No        |
| 17  | I062/136         | Measured Flight Level              | Fixed     | 2      | No        |
| 18  | I062/130         | Calculated Track Geometric Altitude| Fixed     | 2      | No        |
| 19  | I062/135         | Calculated Track Barometric Altitude | Fixed   | 2      | No        |
| 20  | I062/220         | Calculated Rate of Climb/Descent   | Fixed     | 2      | No        |
| 21  | I062/390         | Flight Plan Related Data           | Compound  | 1+     | No        |
| 22  | I062/270         | Target Size & Orientation          | Extended  | 1+     | No        |
| 23  | I062/300         | Vehicle Fleet Identification       | Fixed     | 1      | No        |
| 24  | I062/110         | Mode 5 Data reports & Extended Mode 1 | Compound | 1+    | No        |
| 25  | I062/120         | Track Mode 2 Code                  | Fixed     | 2      | No        |
| 26  | I062/510         | Composed Track Number              | Extended  | 3+     | No        |
| 27  | I062/500         | Estimated Accuracies               | Compound  | 1+     | No        |
| 28  | I062/340         | Measured Information               | Compound  | 1+     | No        |
| 29  | -                | Spare                              | -         | -      | No        |
| 30  | -                | Spare                              | -         | -      | No        |
| 31  | -                | Spare                              | -         | -      | No        |
| 32  | -                | Spare                              | -         | -      | No        |
| 33  | -                | Spare                              | -         | -      | No        |
| 34  | RE062            | Reserved Expansion Field           | Repetitive| 1+     | No        |
| 35  | SP062            | Special Purpose Field              | Repetitive| 1+     | No        |

## Usage

Here's a basic example showing how to create and encode a Category 062 message:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat062"
    v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
    common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func main() {
    // Create a UAP for version 1.20
    uap, err := cat062.NewUAP(cat062.Version120)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        os.Exit(1)
    }
    
    // Create a new data block
    db, err := asterix.NewDataBlock(asterix.Cat062, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create data block: %v\n", err)
        os.Exit(1)
    }
    
    // Create a record
    record, err := asterix.NewRecord(asterix.Cat062, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create record: %v\n", err)
        os.Exit(1)
    }
    
    // Set mandatory items
    record.SetDataItem("I062/010", &common.DataSourceIdentifier{SAC: 25, SIC: 10})
    record.SetDataItem("I062/040", &v120.TrackNumber{Value: 1234})
    
    // Set time of track information
    record.SetDataItem("I062/070", &v120.TimeOfTrackInformation{
        Time: 36000.5, // 10:00:00.5
    })
    
    // Set track status
    trackStatus := &v120.TrackStatus{
        MON: false, // Multisensor track
        SPI: false,
        MRH: true,  // MRH = Geometric altitude used
        SRC: 1,     // GNSS
        CNF: false, // Confirmed track
    }
    trackStatus.SetHasExtension() // Updates internal extension counter
    record.SetDataItem("I062/080", trackStatus)
    
    // Add position in WGS-84
    record.SetDataItem("I062/105", &v120.CalculatedPositionWGS84{
        Latitude:  51.5074, // London
        Longitude: -0.1278,
    })
    
    // Add velocity
    record.SetDataItem("I062/185", &v120.CalculatedTrackVelocity{
        Vx: 100.0, // East component
        Vy: 0.0,   // North component
    })
    
    // Add Mode-3/A code
    record.SetDataItem("I062/060", &v120.TrackMode3ACode{
        Code: 01234, // Octal
    })
    
    // Add target identification
    record.SetDataItem("I062/245", &v120.TargetIdentification{
        IdentType: v120.CallsignRegistration, // Callsign
        Ident: "BAW123",
    })
    
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
    
    fmt.Printf("Successfully encoded Category 062 message (%d bytes)\n", len(encoded))
}
```

## Decoding Example

Here's an example of decoding a received Category 062 message:

```go
func decodeMessage(data []byte) {
    // Create a UAP for version 1.20
    uap, err := cat062.NewUAP(cat062.Version120)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        return
    }
    
    // Create an empty data block
    db, err := asterix.NewDataBlock(asterix.Cat062, uap)
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
        if dataSource, _, exists := record.GetDataItem("I062/010"); exists {
            dsi := dataSource.(*common.DataSourceIdentifier)
            fmt.Printf("  Data Source: SAC=%d, SIC=%d\n", dsi.SAC, dsi.SIC)
        }
        
        // Access track number
        if trackNumber, _, exists := record.GetDataItem("I062/040"); exists {
            tn := trackNumber.(*v120.TrackNumber)
            fmt.Printf("  Track Number: %d\n", tn.Value)
        }
        
        // Access time of track information
        if timeOfTrack, _, exists := record.GetDataItem("I062/070"); exists {
            tot := timeOfTrack.(*v120.TimeOfTrackInformation)
            fmt.Printf("  Time: %s\n", tot.String())
        }
        
        // Access track status
        if trackStatus, _, exists := record.GetDataItem("I062/080"); exists {
            ts := trackStatus.(*v120.TrackStatus)
            fmt.Printf("  Status: %s\n", ts.String())
        }
        
        // Access position
        if position, _, exists := record.GetDataItem("I062/105"); exists {
            pos := position.(*v120.CalculatedPositionWGS84)
            fmt.Printf("  Position: Lat=%.6f°, Lon=%.6f°\n", pos.Latitude, pos.Longitude)
        }
        
        // Access velocity
        if velocity, _, exists := record.GetDataItem("I062/185"); exists {
            vel := velocity.(*v120.CalculatedTrackVelocity)
            fmt.Printf("  Velocity: %s\n", vel.String())
        }
        
        // Access Mode 3/A code
        if mode3A, _, exists := record.GetDataItem("I062/060"); exists {
            m3a := mode3A.(*v120.TrackMode3ACode)
            fmt.Printf("  Mode 3/A: %s\n", m3a.String())
        }
        
        // Access target identification
        if targetId, _, exists := record.GetDataItem("I062/245"); exists {
            ti := targetId.(*v120.TargetIdentification)
            fmt.Printf("  Identification: %s\n", ti.String())
        }
    }
}
```

## Track Status Information

The Track Status data item (I062/080) provides extensive information about the quality and source of a track, including:

- Whether it's a mono-sensor or multi-sensor track
- Source of calculated track altitude
- Track confirmation status
- Coasting indicators for different sensor types
- Military track indicators
- Mode of movement indicators

## Flight Plan Data

The Flight Plan Related Data item (I062/390) can include:

- FPPS identification
- Callsign
- Aircraft type
- Wake turbulence category
- Departure/destination airports
- Assigned SSR code
- Flight status and rules
- Various route information

## Data Ages

The Track Data Ages item (I062/295) provides the age of the data for various sources and data types, allowing the receiving system to determine the freshness of each piece of information.