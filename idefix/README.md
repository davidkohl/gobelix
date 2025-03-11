# Idefix - ASTERIX Message CLI

Idefix is a command-line utility for capturing, decoding, and analyzing ASTERIX (All-purpose STructured Eurocontrol SurveIllance Information EXchange) messages from network traffic. It works with the Gobelix ASTERIX decoding library to provide human-readable output of radar data.

## Features

- Listen on UDP or TCP ports for ASTERIX data
- Filter messages by ASTERIX category (021, 062)
- Print decoded messages in a readable format
- Output to stdout or file

## Installation

### From Source

First, ensure you have Go installed (version 1.19 or later recommended), then:

```bash
go install github.com/davidkohl/idefix@latest
```

Or clone the repository and build:

```bash
git clone https://github.com/davidkohl/idefix.git
cd idefix
go build -o idefix main.go
```

## Usage

### Basic Usage

Listen for all ASTERIX categories on a UDP port:

```bash
idefix dump -p 2000/udp --dumpAll
```

Listen for only Category 021 messages on a TCP port:

```bash
idefix dump -p 62000/tcp --dump021
```

Write output to a file:

```bash
idefix dump -p 2000/udp --dump021 --output asterix_data.txt
```

Enable verbose mode (for debugging):

```bash
idefix dump -p 2000/udp --dumpAll -v
```

### Command Flags

```
  -p, --port string     Port to listen on with protocol (e.g., 2000/udp)
  -o, --output string   Output file (default: stdout)
      --dumpAll         Dump all ASTERIX categories
      --dump021         Dump ASTERIX category 021
      --dump062         Dump ASTERIX category 062
  -v, --verbose         Enable verbose output
```

## Message Format

Idefix outputs decoded ASTERIX messages in a human-readable format. Each message includes:

- ASTERIX category number
- Message size in bytes
- Number of records
- Timestamp of reception
- Source identifier (if available)

For each record, Idefix displays data items in FRN (Field Reference Number) order with their descriptions:

```
ASTERIX CAT021 Message (145 bytes, 1 records)
Timestamp: 2023-05-15T14:23:45.123456Z
Record #1:
  I021/010 (Data Source Identification): SAC=25, SIC=201
  I021/040 (Target Report Descriptor): ATP: 24-Bit ICAO address, ARC: 25ft
  I021/080 (Target Address): AABBCC
  I021/130 (Position in WGS-84 co-ordinates): Lat=49.123, Lon=13.456
  I021/170 (Target Identification): ABC123
```

## Examples

### Monitoring Mode S/ADS-B Traffic

To capture and decode ADS-B messages (typically ASTERIX Category 021):

```bash
idefix dump -p 10001/udp --dump021
```

### Multi-Category Surveillance Data

To capture traffic from a radar system that outputs multiple ASTERIX categories:

```bash
idefix dump -p 8600/tcp --dump021 --dump062
```

### Saving ASTERIX Data for Analysis

To capture data for offline analysis:

```bash
idefix dump -p 2000/udp --dumpAll --output asterix_records.txt
```

## Dependencies

Idefix uses the following libraries:

- [Gobelix](https://github.com/davidkohl/gobelix) - ASTERIX decoding package
- [Cobra](https://github.com/spf13/cobra) - Command-line interface framework


## Acknowledgments

- Named after the faithful companion of Obelix in the Asterix comic series
- Built to complement the Gobelix ASTERIX decoder package