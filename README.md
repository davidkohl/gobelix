# Gobelix - Go ASTERIX Protocol Library

Gobelix is a high-performance Go library for decoding and encoding ASTERIX (All-purpose STructured Eurocontrol SurveIllance Information EXchange) protocol messages. Used in air traffic control, ASTERIX is a binary protocol for exchanging surveillance data.

## Features

- Full support for ASTERIX categories 021, 062
- Efficient binary encoding and decoding
- Streaming support for TCP connections
- Datagram support for UDP
- Thread-safe message processing
- Error handling with detailed context
- Validation of message structure and content
- Extensible architecture for adding new categories

## Installation

```bash
go get -u github.com/davidkohl/gobelix
```

## ASTERIX Protocol Components

Gobelix implements the ASTERIX protocol with the following key components:

### UAP (User Application Profile)

The UAP defines the structure for a specific ASTERIX category, determining:
- Which data items can appear in a message
- The order of data items (Field Reference Numbers)
- The format, type, and optional/mandatory status of each field

Each ASTERIX category has its own UAP implementation.

### Data Items

Data items represent individual fields within an ASTERIX message. Each data item:
- Implements encoding/decoding of its specific binary format
- Provides validation of its contents
- Contains the actual surveillance data (position, identification, etc.)

### FSPEC (Field Specification)

The FSPEC is a bit map at the beginning of each record indicating:
- Which data items are present in the record
- Uses Field Reference Numbers (FRNs) to identify fields
- Extension mechanism for handling many fields

### Record

A record represents data about a single target, containing:
- FSPEC indicating which fields are present
- Collection of data items for that target
- Tied to a specific category and UAP

### Data Block

A data block is a complete ASTERIX message for a single category:
- Contains one or more records
- Has a category identifier and length
- Forms the complete binary message

### ASTERIX Message

The AsterixMessage is the decoded representation of a DataBlock:
- Contains the category, timestamp, and source information
- Holds all decoded records and their data items
- Provides methods to access and query the decoded data

## Core Components

### Decoder

The Decoder is responsible for converting binary ASTERIX data to structured Go objects:
- Takes raw bytes and produces AsterixMessage objects
- Uses the appropriate UAP for each category
- Performs validation and error handling

### Encoder

The Encoder converts structured data into ASTERIX binary format:
- Takes Records or DataBlocks and produces binary data
- Ensures compliance with ASTERIX format specifications
- Handles field ordering and FSPEC generation

### Reader

The Reader provides a streaming interface for ASTERIX data:
- Reads from any io.Reader (TCP connections, files, etc.)
- Handles message framing and boundaries
- Buffers data as needed for complete messages

## Message Structure

An ASTERIX message has the following binary structure:

```
+----------+--------+----------+---------+--------+
| Category | Length | Record 1 | Record 2| ...    |
| (1 byte) |(2 bytes)| (var)   | (var)   | ...    |
+----------+--------+----------+---------+--------+

Each Record:
+-------+----------+----------+----------+
| FSPEC | Data     | Data     | ...      |
|       | Item 1   | Item 2   |          |
+-------+----------+----------+----------+
```

## Error Handling

Gobelix provides robust error handling with context, using custom error types for different situations:

- `ErrInvalidMessage`: General message structure errors
- `ErrInvalidLength`: Length field doesn't match actual data
- `ErrInvalidFSPEC`: Problems with the field specification
- `ErrUnknownCategory`: Unknown ASTERIX category
- `ErrUnknownDataItem`: Unknown data item identifier
- `DecodeError`: Detailed context for decoding failures

## CLI Companion

For a command-line interface to capture and analyze ASTERIX data, check out [Idefix](https://github.com/davidkohl/geobelix/idefix), a companion CLI tool based on Gobelix.

## Examples

For complete usage examples, see the [examples](./examples) directory:

- Basic decoding from a network source
- Creating and encoding ASTERIX messages
- Working with specific categories
- Error handling patterns

## Acknowledgments

Named after the character from the Asterix comic series, with a nod to the Go language.