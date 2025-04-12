# ASTERIX Category 048 - Surveillance Radar and System Track Data

This package implements ASTERIX Category 048 (Surveillance Radar and System Track Data) according to the EUROCONTROL specification.

## Purpose

Category 048 is used to transmit target reports from surveillance radar processors and system track information. It combines plot data from Primary Surveillance Radar (PSR) and Secondary Surveillance Radar (SSR), along with calculated information including:

- Target position (in slant polar and/or Cartesian coordinates)
- Height information
- Mode 3/A codes
- Aircraft identification
- Plot and track quality indicators
- Radar processor information
- Alert messages

This category was historically one of the first ASTERIX formats and is still widely used in radar data distribution systems.

## Implementation

The implementation follows the ASTERIX Category 048 specification with support for different versions. It includes:

- Complete User Application Profile (UAP) definition
- All mandatory and optional data items
- Support for both plot and track data
- Proper radar measurement unit conversions

## Data Items

The following data items are implemented:

| FRN | Data Item        | Description                        | Format    | Length | Mandatory |
|-----|------------------|------------------------------------|-----------|--------|-----------|
| 1   | I048/010         | Data Source Identifier             | Fixed     | 2      | Yes       |
| 2   | I048/140         | Time of Day                        | Fixed     | 3      | Yes       |
| 3   | I048/020         | Target Report Descriptor           | Extended  | 1+     | Yes       |
| 4   | I048/040         | Measured Position                  | Fixed     | 4      | Yes       |
| 5   | I048/070         | Mode-3/A Code                      | Fixed     | 2      | No        |
| 6   | I048/090         | Flight Level                       | Fixed     | 2      | No        |
| 7   | I048/130         | Radar Plot Characteristics         | Extended  | 1+     | No        |
| 8   | I048/220         | Aircraft Address                   | Fixed     | 3      | No        |
| 9   | I048/240         | Aircraft Identification            | Fixed     | 6      | No        |
| 10  | I048/250         | Mode S MB Data                     | Repetitive| 1+     | No        |
| 11  | I048/161         | Track Number                       | Fixed     | 2      | No        |
| 12  | I048/042         | Calculated Position (Cartesian)    | Fixed     | 4      | No        |
| 13  | I048/200         | Calculated Track Velocity          | Fixed     | 4      | No        |
| 14  | I048/170         | Track Status                       | Extended  | 1+     | No        |
| 15  | I048/210         | Track Quality                      | Fixed     | 4      | No        |
| 16  | I048/030         | Warning/Error Conditions           | Extended  | 1+     | No        |
| 17  | I048/080         | Mode-3/A Code Confidence Indicator | Fixed     | 2      | No        |
| 18  | I048/100         | Mode-C Code and Confidence Indicator | Fixed   | 4      | No        |
| 19  | I048/110         | Height Measured by 3D Radar        | Fixed     | 2      | No        |
| 20  | I048/120         | Radial Doppler Speed               | Fixed     | 2      | No        |
| 21  | I048/230         | Communications/ACAS Capability and Flight Status | Fixed | 2 | No    |
| 22  | I048/260         | ACAS Resolution Advisory Report    | Fixed     | 7      | No        |
| 23  | I048/055         | Mode-1 Code                        | Fixed     | 1      | No        |
| 24  | I048/050         | Mode-2 Code                        | Fixed     | 2      | No        |
| 25  | I048/065         | Mode-1 Code Confidence Indicator   | Fixed     | 1      | No        |
| 26  | I048/060         | Mode-2 Code Confidence Indicator   | Fixed     | 2      | No        |
| 27  | RE048            | Reserved Expansion Field           | Repetitive| 1+     | No        |
| 28  | SP048            | Special Purpose Field              | Repetitive| 1+     | No        |

## Usage

Here's a basic example showing how to create and encode a Category 048 message:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat048"
    v17 "github.com/davidkohl/gobelix/cat/cat048/dataitems/v17"
    common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func main() {
    // Create a UAP
    uap, err := cat048.NewUAP(cat048.LatestVersion())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        os.Exit(1)
    }
    
    // Create a new data block
    db, err := asterix.NewDataBlock(asterix.Cat048, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create data block: %v\n", err)
        os.Exit(1)
    }
    
    // Create a record
    record, err := asterix.NewRecord(asterix.Cat048, uap)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create record: %v\n", err)
        os.Exit(1)
    }
    
    // Set mandatory items
    record.SetDataItem("I048/010", &common.DataSourceIdentifier{
        SAC: 25, 
        SIC: 10,
    })
    
    record.SetDataItem("I048/140", &v17.TimeOfDay{
        Time: 36000.5, // 10:00:00.5
    })
    
    // Create target report descriptor
    trd := &v17.TargetReportDescriptor{
        TYP: 5, // PSR + SSR plot
        SIM: false,
        RDP: 0,
        SPI: false,
        RAB: false,
    }
    record.SetDataItem("I048/020", trd)
    
    // Set measured position (range/azimuth)
    record.SetDataItem("I048/040", &v17.MeasuredPosition{
        Range:   150.5, // NM
        Azimuth: 270.5, // degrees
    })
    
    // Add Mode 3/A code
    record.SetDataItem("I048/070", &v17.Mode3ACode{
        Code:      4321, // Octal
        Validated: true,
        Garbled:   false,
    })
    
    // Add flight level
    record.SetDataItem("I048/090", &v17.FlightLevel{
        Value:     350.0, // Flight Level 350
        Validated: true,
        Garbled:   false,
    })
    
    // Add track number (for system tracks)
    record.SetDataItem("I048/161", &v17.TrackNumber{
        Value: 1234,
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
    
    fmt.Printf("Successfully encoded Category 048 message (%d bytes)\n", len(encoded))
}
```

## Decoding Example

Here's an example of decoding a received Category 048 message:

```go
func decodeMessage(data []byte) {
    // Create a UAP
    uap, err := cat048.NewUAP(cat048.LatestVersion())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create UAP: %v\n", err)
        return
    }
    
    // Create an empty data block
    db, err := asterix.NewDataBlock(asterix.Cat048, uap)
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
        if dataSource, _, exists := record.GetDataItem("I048/010"); exists {
            dsi := dataSource.(*common.DataSourceIdentifier)
            fmt.Printf("  Data Source: SAC=%d, SIC=%d\n", dsi.SAC, dsi.SIC)
        }
        
        // Access time of day
        if timeOfDay, _, exists := record.GetDataItem("I048/140"); exists {
            tod := timeOfDay.(*v17.TimeOfDay)
            fmt.Printf("  Time: %s\n", tod.String())
        }
        
        // Access target report descriptor
        if trd, _, exists := record.GetDataItem("I048/020"); exists {
            descriptor := trd.(*v17.TargetReportDescriptor)
            fmt.Printf("  Report Type: %s\n", descriptor.String())
        }
        
        // Access position
        if position, _, exists := record.GetDataItem("I048/040"); exists {
            pos := position.(*v17.MeasuredPosition)
            fmt.Printf("  Position: Range=%.2f NM, Azimuth=%.2f°\n", 
                pos.Range, pos.Azimuth)
        }
        
        // Access Mode 3/A code
        if mode3A, _, exists := record.GetDataItem("I048/070"); exists {
            m3a := mode3A.(*v17.Mode3ACode)
            fmt.Printf("  Mode 3/A: %04o\n", m3a.Code)
        }
        
        // Access flight level
        if flightLevel, _, exists := record.GetDataItem("I048/090"); exists {
            fl := flightLevel.(*v17.FlightLevel)
            fmt.Printf("  Flight Level: FL%.0f\n", fl.Value)
        }
        
        // Access track number
        if trackNum, _, exists := record.GetDataItem("I048/161"); exists {
            tn := trackNum.(*v17.TrackNumber)
            fmt.Printf("  Track Number: %d\n", tn.Value)
        }
    }
}
```

## Report Types

Category 048 can contain various types of reports as indicated by the TYP field in the Target Report Descriptor:

- PSR plot
- SSR plot
- Combined PSR+SSR plot
- Mode S surveillance responses
- System tracks
- Various warning/error conditions

## Measurement Units

Key measurement units used in Category 048:

- Range: 1/256 NM (1/256 nautical miles)
- Azimuth: 360°/65536 (approximately 0.0055°)
- Flight Level: 1/4 FL (25 feet)
- Speeds: Various scales depending on context
- Time: 1/128 seconds since midnight UTC

## Warning and Error Conditions

The Warning/Error Conditions data item (I048/030) provides information about failures or error conditions detected by the surveillance radar data processing system, including:

- Distorted plot data
- Ghost plots/tracks
- Possible split tracks
- Chain detection issues
- Overloads

## Aircraft Identification

For Mode S equipped aircraft, the Aircraft Identification data item (I048/240) provides the callsign or registration mark of the aircraft.