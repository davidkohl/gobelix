# ASTERIX Category 020 - Multilateration Target Reports

This package implements ASTERIX Category 020 (Multilateration Target Reports) according to the EUROCONTROL specification.

## Purpose

Category 020 is used for the transmission of target reports from multilateration systems. These systems use multiple ground stations to determine aircraft position through time-difference-of-arrival (TDOA) measurements. The category provides:

- Target position in WGS-84 coordinates
- Target identification
- Mode 3/A and Mode S codes
- Vehicle fleet identification for surface targets
- Contributing devices information
- Track status and quality

## Implementation

The implementation follows the ASTERIX Category 020 specification version 1.5. It includes:

- Complete User Application Profile (UAP) definition
- All mandatory and optional data items
- Support for both airborne and surface targets

## Data Items

| FRN | Data Item  | Description                        | Format   | Length | Mandatory |
|-----|------------|------------------------------------|----------|--------|-----------|
| 1   | I020/010   | Data Source Identifier             | Fixed    | 2      | Yes       |
| 2   | I020/020   | Target Report Descriptor           | Extended | 1+     | Yes       |
| 3   | I020/140   | Time of Day                        | Fixed    | 3      | Yes       |
| 4   | I020/041   | Position in WGS-84                 | Fixed    | 8      | No        |
| 5   | I020/042   | Position in Cartesian Coordinates  | Fixed    | 6      | No        |
| 6   | I020/161   | Track Number                       | Fixed    | 2      | No        |
| 7   | I020/170   | Track Status                       | Extended | 1+     | No        |
| 8   | I020/070   | Mode-3/A Code                      | Fixed    | 2      | No        |
| 9   | I020/202   | Calculated Track Velocity          | Fixed    | 4      | No        |
| 10  | I020/090   | Flight Level                       | Fixed    | 2      | No        |
| 11  | I020/100   | Mode-C Code                        | Fixed    | 4      | No        |
| 12  | I020/220   | Target Address                     | Fixed    | 3      | No        |
| 13  | I020/245   | Target Identification              | Fixed    | 7      | No        |
| 14  | I020/110   | Measured Height                    | Fixed    | 2      | No        |
| 15  | I020/105   | Geometric Height                   | Fixed    | 2      | No        |
| 16  | I020/210   | Calculated Acceleration            | Fixed    | 2      | No        |
| 17  | I020/300   | Vehicle Fleet Identification       | Fixed    | 1      | No        |
| 18  | I020/310   | Pre-programmed Message             | Fixed    | 1      | No        |
| 19  | I020/500   | Position Accuracy                  | Compound | 1+     | No        |
| 20  | I020/400   | Contributing Devices               | Repetitive | 1+   | No        |

## Usage

```go
import (
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat020"
)

// Create a UAP
uap, err := cat020.NewUAP(cat020.LatestVersion())
if err != nil {
    // handle error
}

// Create a data block
db, err := asterix.NewDataBlock(asterix.Cat020, uap)
if err != nil {
    // handle error
}
```

## Measurement Units

- Position: WGS-84 latitude/longitude in 180°/2^25 degrees
- Height: 25 feet
- Flight Level: 1/4 FL (25 feet)
- Speed: 0.25 m/s
- Time: 1/128 seconds since midnight UTC
