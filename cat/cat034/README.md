# ASTERIX Category 034 - Monoradar Service Messages

This package implements ASTERIX Category 034 (Monoradar Service Messages) according to the EUROCONTROL specification.

## Purpose

Category 034 is used for transmission of status and configuration messages from monoradar data sources. It provides comprehensive information about the radar system, including:

- Sector crossing and north/south marker messages
- System configuration and processing status
- Antenna rotation period
- Message count statistics
- Collimation and bias corrections

## Implementation

The implementation follows the ASTERIX Category 034 specification version 1.29. It includes:

- Complete User Application Profile (UAP) definition
- All mandatory and optional data items
- System status monitoring support

## Data Items

| FRN | Data Item  | Description                     | Format   | Length | Mandatory |
|-----|------------|---------------------------------|----------|--------|-----------|
| 1   | I034/010   | Data Source Identifier          | Fixed    | 2      | Yes       |
| 2   | I034/000   | Message Type                    | Fixed    | 1      | Yes       |
| 3   | I034/030   | Time of Day                     | Fixed    | 3      | Yes       |
| 4   | I034/020   | Sector Number                   | Fixed    | 1      | Yes       |
| 5   | I034/041   | Antenna Rotation Period         | Fixed    | 2      | No        |
| 6   | I034/050   | System Configuration Status     | Compound | 1+     | No        |
| 7   | I034/060   | System Processing Mode          | Compound | 1+     | No        |
| 8   | I034/070   | Message Count Values            | Repetitive | 1+   | No        |
| 9   | I034/100   | Generic Polar Window            | Fixed    | 8      | No        |
| 10  | I034/110   | Data Filter                     | Fixed    | 1      | No        |
| 11  | I034/120   | 3D Position of Data Source      | Fixed    | 8      | No        |
| 12  | I034/090   | Collimation Error               | Fixed    | 2      | No        |

## Message Types

- 1: North Marker message
- 2: Sector Crossing message
- 3: Geographical Filtering message
- 4: Jamming Strobe message

## Usage

```go
import (
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat034"
)

// Create a UAP
uap, err := cat034.NewUAP(cat034.LatestVersion())
if err != nil {
    // handle error
}

// Create a data block
db, err := asterix.NewDataBlock(asterix.Cat034, uap)
if err != nil {
    // handle error
}
```

## Measurement Units

- Sector Azimuth: 360°/256 (approximately 1.41°)
- Antenna Rotation Period: 1/128 seconds
- Time: 1/128 seconds since midnight UTC
- Collimation Error: 0.022° for azimuth, 1/256 NM for range
