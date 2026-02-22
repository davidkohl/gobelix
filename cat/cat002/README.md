# ASTERIX Category 002 - Monoradar Service Messages

This package implements ASTERIX Category 002 (Monoradar Service Messages) according to the EUROCONTROL specification.

## Purpose

Category 002 is used for transmission of monoradar service messages. These messages provide information about the radar system status and operation, including:

- North/South markers (antenna position references)
- Sector crossing messages
- Blind zone filter activation/deactivation
- Station configuration and processing status
- Plot count statistics

## Implementation

The implementation follows the ASTERIX Category 002 specification version 1.0. It includes:

- Complete User Application Profile (UAP) definition
- All mandatory and optional data items
- Support for service message types

## Data Items

| FRN | Data Item  | Description                      | Format   | Length | Mandatory |
|-----|------------|----------------------------------|----------|--------|-----------|
| 1   | I002/010   | Data Source Identifier           | Fixed    | 2      | Yes       |
| 2   | I002/000   | Message Type                     | Fixed    | 1      | Yes       |
| 3   | I002/020   | Sector Number                    | Fixed    | 1      | No        |
| 4   | I002/030   | Time of Day                      | Fixed    | 3      | No        |
| 5   | I002/041   | Antenna Rotation Speed           | Fixed    | 2      | No        |
| 6   | I002/050   | Station Configuration Status     | Extended | 1+     | No        |
| 7   | I002/060   | Station Processing Mode          | Extended | 1+     | No        |
| 8   | I002/070   | Plot Count Values                | Repetitive | 1+   | No        |
| 9   | I002/100   | Dynamic Window - Type 1          | Fixed    | 8      | No        |
| 10  | I002/090   | Collimation Error                | Fixed    | 2      | No        |
| 11  | I002/080   | Warning/Error Conditions         | Extended | 1+     | No        |

## Message Types

- 1: North Marker message
- 2: Sector Crossing message
- 3: South Marker message
- 8: Activation of blind zone filter
- 9: Stop of blind zone filter

## Usage

```go
import (
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat002"
)

// Create a UAP
uap, err := cat002.NewUAP(cat002.LatestVersion())
if err != nil {
    // handle error
}

// Create a data block
db, err := asterix.NewDataBlock(asterix.Cat002, uap)
if err != nil {
    // handle error
}
```

## Measurement Units

- Sector Azimuth: 360°/256 (approximately 1.41°)
- Antenna Rotation Period: 1/128 seconds
- Time: 1/128 seconds since midnight UTC
