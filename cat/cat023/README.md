# ASTERIX Category 023 - CNS/ATM Ground Station Service Messages

This package implements ASTERIX Category 023 (CNS/ATM Ground Station Service Messages) according to the EUROCONTROL specification.

## Purpose

Category 023 is used for transmission of service messages from CNS/ATM ground stations. It provides information about the status and performance of ground-based surveillance equipment, including:

- Ground station status and configuration
- Service identification
- Operational range information
- Service statistics
- Alert conditions

## Implementation

The implementation follows the ASTERIX Category 023 specification version 1.3. It includes:

- Complete User Application Profile (UAP) definition
- All mandatory and optional data items
- Service status monitoring support

## Data Items

| FRN | Data Item  | Description                     | Format   | Length | Mandatory |
|-----|------------|---------------------------------|----------|--------|-----------|
| 1   | I023/010   | Data Source Identifier          | Fixed    | 2      | Yes       |
| 2   | I023/000   | Report Type                     | Fixed    | 1      | Yes       |
| 3   | I023/015   | Service Type and Identification | Fixed    | 1      | No        |
| 4   | I023/070   | Time of Day                     | Fixed    | 3      | No        |
| 5   | I023/100   | Ground Station Status           | Extended | 1+     | No        |
| 6   | I023/101   | Service Configuration           | Extended | 1+     | No        |
| 7   | I023/200   | Operational Range               | Fixed    | 1      | No        |
| 8   | I023/110   | Service Status                  | Fixed    | 1      | No        |
| 9   | I023/120   | Service Statistics              | Repetitive | 1+   | No        |

## Report Types

- 1: Ground Station Status report
- 2: Service Status report
- 3: Service Statistics report

## Usage

```go
import (
    "github.com/davidkohl/gobelix/asterix"
    "github.com/davidkohl/gobelix/cat/cat023"
)

// Create a UAP
uap, err := cat023.NewUAP(cat023.LatestVersion())
if err != nil {
    // handle error
}

// Create a data block
db, err := asterix.NewDataBlock(asterix.Cat023, uap)
if err != nil {
    // handle error
}
```

## Measurement Units

- Time: 1/128 seconds since midnight UTC
- Operational Range: 1 NM (nautical mile)
